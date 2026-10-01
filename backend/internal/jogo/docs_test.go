package jogo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// raizDoProjeto é o caminho da raiz do repositório a partir deste pacote.
var raizDoProjeto = filepath.Join("..", "..", "..")

// blocoJSON devolve o primeiro bloco ```json da seção de docs/<arquivo> cujo
// título começa com prefixoTitulo (ex.: "## 2."). Falha com mensagem clara se
// a seção ou o bloco não existirem.
func blocoJSON(t *testing.T, arquivo, prefixoTitulo string) []byte {
	t.Helper()
	dados, err := os.ReadFile(filepath.Join(raizDoProjeto, "docs", arquivo))
	if err != nil {
		t.Fatal(err)
	}
	nivel := strings.SplitN(prefixoTitulo, " ", 2)[0] + " "
	linhas := strings.Split(string(dados), "\n")
	dentroDaSecao, dentroDoBloco := false, false
	var bloco []string
	for _, linha := range linhas {
		switch {
		case dentroDoBloco:
			if strings.HasPrefix(linha, "```") {
				return []byte(strings.Join(bloco, "\n"))
			}
			bloco = append(bloco, linha)
		case strings.HasPrefix(linha, prefixoTitulo):
			dentroDaSecao = true
		case dentroDaSecao && strings.HasPrefix(linha, nivel):
			dentroDaSecao = false
		case dentroDaSecao && strings.HasPrefix(linha, "```json"):
			dentroDoBloco = true
		}
	}
	t.Fatalf("bloco json da seção %q de docs/%s não encontrado", prefixoTitulo, arquivo)
	return nil
}

// idaEVolta decodifica dados em destino, codifica de novo e exige que o
// resultado seja semanticamente igual ao original (D3).
func idaEVolta[T any](t *testing.T, dados []byte) {
	t.Helper()
	var valor T
	if err := json.Unmarshal(dados, &valor); err != nil {
		t.Fatalf("decodificar: %v", err)
	}
	codificado, err := json.Marshal(valor)
	if err != nil {
		t.Fatalf("codificar: %v", err)
	}
	var esperado, obtido any
	if err := json.Unmarshal(dados, &esperado); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(codificado, &obtido); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(esperado, obtido) {
		t.Errorf("ida e volta diferente\n esperado: %s\n obtido:   %s", dados, codificado)
	}
}

// chaves devolve as chaves do objeto JSON gerado ao codificar v.
func chaves(t *testing.T, v any) map[string]any {
	t.Helper()
	codificado, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(codificado, &m); err != nil {
		t.Fatal(err)
	}
	return m
}
