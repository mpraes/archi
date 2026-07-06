package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mpraes/archi/internal/model"
)

func newExportCmd(g *GlobalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "export [caminho_do_projeto]",
		Short: "Exporta as métricas do projeto (JSON/Markdown)",
		Long: "Gera um relatório estático das métricas sem abrir o servidor local.\n" +
			"Formatos suportados: json, markdown.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := g.pathArg(args)
			summary, warnings := scan(path, g)
			for _, w := range warnings {
				fmt.Fprintln(os.Stderr, "# warning:", w)
			}
			switch strings.ToLower(format) {
			case "json":
				return exportJSON(os.Stdout, summary)
			case "markdown", "md":
				return exportMarkdown(os.Stdout, summary)
			default:
				return fmt.Errorf("formato não suportado: %s (use json ou markdown)", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "Formato de saída: json ou markdown")
	return cmd
}

func exportJSON(w io.Writer, s model.Summary) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func exportMarkdown(w io.Writer, s model.Summary) error {
	if err := writeMarkdownHeader(w, s); err != nil {
		return err
	}
	if err := writeMarkdownModules(w, s.Modules); err != nil {
		return err
	}
	return writeMarkdownConnascence(w, s.Connascence)
}

func writeMarkdownHeader(w io.Writer, s model.Summary) error {
	if _, err := fmt.Fprintf(w, "# Archi — %s\n\n", s.ProjectName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Módulos analisados: **%d**\n\n", s.ModuleCount); err != nil {
		return err
	}
	if len(s.Hotspots) == 0 {
		return nil
	}
	if _, err := fmt.Fprintf(w, "## Hotspots\n\n"); err != nil {
		return err
	}
	for _, h := range s.Hotspots {
		if _, err := fmt.Fprintf(w, "- `%s`\n", h); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w)
	return err
}

func writeMarkdownModules(w io.Writer, modules []model.ModuleMetrics) error {
	lines := []string{
		"## Módulos", "",
		"| Módulo | Ca | Ce | I | A | D | Complexidade máx | Abstratos | Concretos |",
		"|--------|----|----|---|---|---|------------------|-----------|-----------|",
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	for _, m := range modules {
		if _, err := fmt.Fprintf(w, "| `%s` | %d | %d | %.2f | %.2f | %.2f | %d | %d | %d |\n",
			m.Module, m.Afferent, m.Efferent, m.Instability, m.Abstraction,
			m.Distance, m.MaxComplexity, m.Abstracts, m.Concretes); err != nil {
			return err
		}
	}
	return nil
}

func writeMarkdownConnascence(w io.Writer, conns []model.Connascence) error {
	if len(conns) == 0 {
		return nil
	}
	lines := []string{
		"", "## Conascência", "",
		"| Tipo | De | Para | Detalhe |",
		"|------|----|------|---------|",
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	for _, c := range conns {
		if _, err := fmt.Fprintf(w, "| %s | `%s` | `%s` | %s |\n", c.Kind, c.From, c.To, c.Detail); err != nil {
			return err
		}
	}
	return nil
}

var _ = strings.ToLower
