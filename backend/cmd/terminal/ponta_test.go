package main

import (
	"strings"
	"testing"
	"time"
)

// mapaExemplo é o caminho de mapas/exemplo.json visto de cmd/terminal.
const mapaExemplo = "../../../mapas/exemplo.json"

func TestPartidaNoMapaExemplo(t *testing.T) {
	saida, erros, codigo := rodarTeste("-mapa", mapaExemplo, "-semente", "1", "-atraso", "0")
	if codigo != 0 || erros != "" {
		t.Fatalf("CA-08 código %d, erros %q", codigo, erros)
	}
	if !strings.HasPrefix(ultimaLinha(saida), "Fim: ") {
		t.Errorf("FIM-02 CA-08 última linha %q", ultimaLinha(saida))
	}
	if _, turno := cabecalhos(saida); turno < 1 || turno > 50 {
		t.Errorf("DEC-07 CA-08 maior turno impresso %d, limite 50", turno)
	}
}

func TestMapaPadrao(t *testing.T) {
	t.Chdir("../..") // backend/: o padrão encontra ../mapas/exemplo.json (D8)
	saida, erros, codigo := rodarTeste("-atraso", "0")
	if codigo != 0 || !strings.HasPrefix(ultimaLinha(saida), "Fim: ") {
		t.Errorf("MAP-03 código %d, erros %q", codigo, erros)
	}
}

func TestSaidaDeterministica(t *testing.T) {
	a, _, _ := rodarTeste("-mapa", mapaExemplo, "-semente", "1", "-atraso", "0")
	b, _, _ := rodarTeste("-mapa", mapaExemplo, "-semente", "1", "-atraso", "0")
	c, _, _ := rodarTeste("-mapa", mapaExemplo, "-semente", "2", "-atraso", "0")
	if a != b {
		t.Error("CA-13 mesma semente, saídas diferentes")
	}
	if a == c {
		t.Error("CA-13 sementes 1 e 2, saídas iguais")
	}
}

func TestTabuleirosImpressos(t *testing.T) {
	for _, semente := range []uint64{1, 2, 3} {
		estado, jogadores, err := preparar(mapaExemplo, "aleatorio-v1", semente, catalogoDeTeste())
		if err != nil {
			t.Fatal(err)
		}
		var b strings.Builder
		r := jogar(estado, jogadores, 0, &b)
		saida := b.String()
		etapas, _ := cabecalhos(saida)
		if etapas != r.relatorios {
			t.Errorf("CA-16 semente %d: %d etapas impressas, %d relatórios", semente, etapas, r.relatorios)
		}
		linhas := strings.Split(saida, "\n")
		tabuleiros := 0
		for i, l := range linhas {
			if !strings.HasPrefix(l, "Turno ") {
				continue
			}
			tabuleiros++
			for _, linha := range linhas[i+1 : i+1+estado.Config.Altura] {
				if len(linha) != estado.Config.Largura || strings.Trim(linha, ".#+o*&123456789") != "" {
					t.Fatalf("CA-16 semente %d, %s: linha %q", semente, l, linha)
				}
			}
		}
		if tabuleiros != r.relatorios+1 {
			t.Errorf("CA-16 semente %d: %d tabuleiros, esperado %d (relatórios + início)", semente, tabuleiros, r.relatorios+1)
		}
	}
}

func TestAtraso(t *testing.T) {
	const x = 10 * time.Millisecond
	porArgumento := gravarMapa(t, mapaDe(t, emL))
	doMapa := mapaDe(t, emL)
	doMapa.Config.DuracaoEtapaMs = int(x / time.Millisecond)
	casos := []struct {
		nome string
		args []string
	}{
		{"EST-02 CA-20 atraso por argumento", []string{"-mapa", porArgumento, "-bots", "bombista,espera", "-atraso", x.String()}},
		{"EST-02 CA-20 atraso padrão do mapa", []string{"-mapa", gravarMapa(t, doMapa), "-bots", "bombista,espera"}},
	}
	for _, c := range casos {
		t.Run(c.nome, func(t *testing.T) {
			inicio := time.Now()
			saida, erros, codigo := rodarTeste(c.args...)
			duracao := time.Since(inicio)
			if codigo != 0 {
				t.Fatalf("código %d: %s", codigo, erros)
			}
			etapas, _ := cabecalhos(saida) // tabuleiros − 1
			if minimo := time.Duration(etapas) * x; duracao < minimo {
				t.Errorf("durou %v, mínimo %v para %d etapas", duracao, minimo, etapas)
			}
		})
	}
}
