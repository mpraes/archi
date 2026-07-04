// Package cli implements the archi command surface with cobra.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// GlobalFlags holds flags applying to the scan + server flow.
type GlobalFlags struct {
	AIPresent bool
	APIKey    string
	Port      int
	NoBrowser bool
	Lang      string
	Exclude   []string
}

// NewRootCmd builds the root command with persistent flags shared with subcommands.
func NewRootCmd(version string) *cobra.Command {
	g := &GlobalFlags{Port: 8080}

	root := &cobra.Command{
		Use:   "archi [caminho_do_projeto] [flags]",
		Short: "Archi — análise estática e diagnóstico visual de arquitetura",
		Long: "Archi é uma ferramenta de análise estática que varre um projeto, calcula métricas\n" +
			"arquiteturais (acoplamento, instabilidade, abstração, sequência principal) e serve\n" +
			"uma interface web local embutida para diagnóstico visual.",
		Version: version,
		Args:    cobra.MaximumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error { return runRoot(cmd, args, g) },
	}
	attachStyledHelp(root)
	root.Flags().BoolVarP(&g.AIPresent, "ai", "a", false, "Ativa os insights do consultor virtual via IA")
	root.Flags().StringVar(&g.APIKey, "api-key", "", "Chave de API do Gemini (alternativa à variável de ambiente)")
	root.Flags().IntVarP(&g.Port, "port", "p", 8080, "Porta para o servidor web local")
	root.Flags().BoolVar(&g.NoBrowser, "no-browser", false, "Não abre o navegador automaticamente após o escaneamento")
	root.Flags().StringVarP(&g.Lang, "lang", "l", "", "Força uma linguagem específica para o parsing (go, ts)")
	root.Flags().StringSliceVar(&g.Exclude, "exclude", nil, "Pastas ou arquivos adicionais para ignorar")

	root.AddCommand(newExportCmd(g))
	root.AddCommand(newCheckCmd(g))
	return root
}

// Execute runs the root command and exits with the appropriate code.
func Execute(version string) {
	if err := NewRootCmd(version).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func (g *GlobalFlags) pathArg(args []string) string {
	if len(args) > 0 && args[0] != "" {
		return args[0]
	}
	return "."
}
