package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// exitCodeError signals a non-zero process exit without calling os.Exit in tests.
type exitCodeError struct {
	code int
	msg  string
}

func (e *exitCodeError) Error() string { return e.msg }

func newCheckCmd(g *GlobalFlags) *cobra.Command {
	var maxDistance float64
	cmd := &cobra.Command{
		Use:   "check [caminho_do_projeto]",
		Short: "Valida limites arquiteturais no pipeline de CI",
		Long: "Roda a análise e retorna código de saída de erro caso alguma métrica\n" +
			"arquitetural seja violada. Ideal para CI/CD.",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := g.pathArg(args)
			summary, _ := scan(path, g)

			violations := 0
			for _, m := range summary.Modules {
				if maxDistance > 0 && m.Distance > maxDistance {
					fmt.Fprintf(os.Stderr, "VIOLAÇÃO: módulo `%s` com D=%.3f > --max-distance=%.3f\n",
						m.Module, m.Distance, maxDistance)
					violations++
				}
			}
			if violations == 0 {
				fmt.Printf("OK: %d módulos, todos dentro do limite.\n", summary.ModuleCount)
				return nil
			}
			fmt.Fprintf(os.Stderr, "\n%d violações encontradas.\n", violations)
			return &exitCodeError{code: 1, msg: fmt.Sprintf("%d violações encontradas", violations)}
		},
	}
	cmd.Flags().Float64Var(&maxDistance, "max-distance", 0.8, "Distância máxima da sequência principal permitida")
	return cmd
}
