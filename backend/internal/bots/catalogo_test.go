package bots

import (
	"slices"
	"strings"
	"testing"
)

func TestCatalogoPadrao(t *testing.T) {
	c := Padrao()
	versoes := c.Versoes()
	if !slices.Contains(versoes, "aleatorio-v1") {
		t.Errorf("CA-01 aleatorio-v1 fora do catálogo: %v", versoes)
	}
	if !slices.Contains(versoes, "aleatorio-v2") {
		t.Errorf("API-09 CA-03 (marco 13) aleatorio-v2 fora do catálogo: %v", versoes)
	}
	if !slices.Contains(versoes, "aleatorio-v3") {
		t.Errorf("API-09 CA-03 (marco 15) aleatorio-v3 fora do catálogo: %v", versoes)
	}
	if !slices.Contains(versoes, "aleatorio-v4") {
		t.Errorf("API-09 CA-03 (marco 16) aleatorio-v4 fora do catálogo: %v", versoes)
	}
	if !slices.IsSorted(versoes) || len(slices.Compact(slices.Clone(versoes))) != len(versoes) {
		t.Errorf("CA-01 versões fora de ordem ou repetidas: %v", versoes)
	}
	for _, v := range versoes {
		bot, err := c.Criar(v, 1)
		if err != nil {
			t.Fatalf("CA-01 Criar(%q): %v", v, err)
		}
		if bot.Versao() != v {
			t.Errorf("CA-01 Criar(%q).Versao() = %q", v, bot.Versao())
		}
	}
}

func TestCriarVersaoDesconhecida(t *testing.T) {
	bot, err := Padrao().Criar("inexistente-v9", 1)
	if err == nil || bot != nil {
		t.Fatalf("CA-02 esperado erro, obtido bot %v e erro %v", bot, err)
	}
	if !strings.Contains(err.Error(), "inexistente-v9") {
		t.Errorf("CA-02 erro não cita a versão: %v", err)
	}
}
