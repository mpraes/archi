package cli

import (
	"encoding/json"
	"fmt"
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
				return exportJSON(summary)
			case "markdown", "md":
				return exportMarkdown(summary)
			default:
				return fmt.Errorf("formato não suportado: %s (use json ou markdown)", format)
			}
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "Formato de saída: json ou markdown")
	return cmd
}

func exportJSON(s model.Summary) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func exportMarkdown(s model.Summary) error {
	fmt.Printf("# Archi — %s\n\n", s.ProjectName)
	fmt.Printf("Módulos analisados: **%d**\n\n", s.ModuleCount)
	if len(s.Hotspots) > 0 {
		fmt.Printf("## Hotspots\n\n")
		for _, h := range s.Hotspots {
			fmt.Printf("- `%s`\n", h)
		}
		fmt.Println()
	}
	fmt.Println("## Módulos")
	fmt.Println()
	fmt.Println("| Módulo | Ca | Ce | I | A | D | Complexidade máx | Abstratos | Concretos |")
	fmt.Println("|--------|----|----|---|---|---|------------------|-----------|-----------|")
	for _, m := range s.Modules {
		fmt.Printf("| `%s` | %d | %d | %.2f | %.2f | %.2f | %d | %d | %d |\n",
			m.Module, m.Afferent, m.Efferent, m.Instability, m.Abstraction,
			m.Distance, m.MaxComplexity, m.Abstracts, m.Concretes)
	}
	if len(s.Connascence) > 0 {
		fmt.Println()
		fmt.Println("## Conascência")
		fmt.Println()
		fmt.Println("| Tipo | De | Para | Detalhe |")
		fmt.Println("|------|----|------|---------|")
		for _, c := range s.Connascence {
			fmt.Printf("| %s | `%s` | `%s` | %s |\n", c.Kind, c.From, c.To, c.Detail)
		}
	}
	return nil
}

var _ = strings.ToLower
