package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kustavo/bomber-guess/backend/internal/jogo"
)

// gravarMapa grava o mapa em JSON num diretório temporário e devolve o caminho.
func gravarMapa(t *testing.T, m jogo.Mapa) string {
	t.Helper()
	dados, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	caminho := filepath.Join(t.TempDir(), "mapa.json")
	if err := os.WriteFile(caminho, dados, 0o644); err != nil {
		t.Fatal(err)
	}
	return caminho
}

// rodarTeste roda o terminal com o catálogo de teste.
func rodarTeste(args ...string) (saida, erros string, codigo int) {
	var s, e strings.Builder
	codigo = rodar(args, &s, &e, catalogoDeTeste())
	return s.String(), e.String(), codigo
}

func TestListar(t *testing.T) {
	saida, erros, codigo := rodarTeste("-listar")
	esperado := strings.Join(catalogoDeTeste().Versoes(), "\n") + "\n"
	if codigo != 0 || saida != esperado || erros != "" {
		t.Errorf("CA-03 código %d, saída %q, erros %q", codigo, saida, erros)
	}
}

func TestErrosDeUso(t *testing.T) {
	invalido := mapaDe(t, "1.2")
	invalido.JogadorPadrao.AcoesPorTurno = 0
	dois := gravarMapa(t, mapaDe(t, "1.2"))
	casos := []struct {
		nome   string
		args   []string
		motivo string
	}{
		{"MAP-05 CA-17 mapa inexistente", []string{"-mapa", "nao/existe.json"}, "nao/existe.json"},
		{"MAP-05 CA-17 mapa inválido", []string{"-mapa", gravarMapa(t, invalido)}, "acoes_por_turno"},
		{"MAP-06 CA-18 versão desconhecida", []string{"-mapa", dois, "-bots", "espera,nenhum-v9"}, "nenhum-v9"},
		{"MAP-06 CA-18 quantidade de versões errada", []string{"-mapa", dois, "-bots", "espera,espera,espera"}, "3 versões"},
		{"CA-18 mais de 9 jogadores", []string{"-mapa", gravarMapa(t, mapaDe(t, "1.2\n...\n...\n...", func(m *jogo.Mapa) {
			for y := 1; y <= 3; y++ { // mais 9 posições: 11 no total
				for x := range 3 {
					m.PosicoesIniciais = append(m.PosicoesIniciais, jogo.Posicao{X: x, Y: y})
				}
			}
		}))}, "no máximo 9"},
		{"CA-18 argumento desconhecido", []string{"-cor"}, "-cor"},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			saida, erros, codigo := rodarTeste(append(c.args, "-atraso", "0")...)
			if codigo != 2 || saida != "" || !strings.Contains(erros, c.motivo) {
				t.Errorf("código %d, saída %q, erros %q (esperado motivo %q)", codigo, saida, erros, c.motivo)
			}
		})
	}
}

func TestVersoesPorPosicao(t *testing.T) {
	mapa := gravarMapa(t, mapaDe(t, "1..\n...\n..2", comLimite(1)))
	casos := []struct {
		nome       string
		bots       string
		infratores []string
	}{
		{"MAP-03 CA-19 uma versão joga em todas as posições", "fora", []string{"jogador_1", "jogador_2"}},
		{"MAP-03 CA-19 versão i na posição i", "fora,espera", []string{"jogador_1"}},
		{"MAP-03 CA-19 versão i na posição i, invertido", "espera,fora", []string{"jogador_2"}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			saida, erros, codigo := rodarTeste("-mapa", mapa, "-bots", c.bots, "-atraso", "0")
			if codigo != 0 {
				t.Fatalf("código %d: %s", codigo, erros)
			}
			for _, id := range []string{"jogador_1", "jogador_2"} {
				infringiu := strings.Contains(saida, "infração: "+id+",")
				esperado := strings.Contains(strings.Join(c.infratores, " "), id)
				if infringiu != esperado {
					t.Errorf("%s: infração %v, esperado %v\n%s", id, infringiu, esperado, saida)
				}
			}
		})
	}
}
