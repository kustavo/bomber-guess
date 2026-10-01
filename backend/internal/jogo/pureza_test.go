package jogo

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestPureza verifica que o código de produção do pacote importa apenas a
// biblioteca padrão (CA-12).
func TestPureza(t *testing.T) {
	arquivos, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	conjunto := token.NewFileSet()
	analisados := 0
	for _, arquivo := range arquivos {
		if strings.HasSuffix(arquivo, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(conjunto, arquivo, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		analisados++
		for _, imp := range f.Imports {
			caminho, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			primeiro, _, _ := strings.Cut(caminho, "/")
			if strings.Contains(primeiro, ".") {
				t.Errorf("CA-12 %s importa %q, que não é da biblioteca padrão", arquivo, caminho)
			}
		}
	}
	if analisados == 0 {
		t.Fatal("CA-12 nenhum arquivo de produção encontrado")
	}
}
