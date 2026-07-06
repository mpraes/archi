package cli

import (
	"bytes"
	"testing"
)

func TestStyledHelp(t *testing.T) {
	root := NewRootCmd("test")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("empty help output")
	}
}

func TestExportHelp(t *testing.T) {
	root := NewRootCmd("test")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"export", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("export")) {
		t.Fatal("missing export in help")
	}
}

func TestAIEnabled(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")
	g := &GlobalFlags{AIPresent: true}
	if aiEnabled(g) == nil {
		t.Fatal("expected AI config")
	}
	t.Setenv("GEMINI_API_KEY", "")
	g2 := &GlobalFlags{}
	if aiEnabled(g2) != nil {
		t.Fatal("expected nil without flag/key")
	}
}
