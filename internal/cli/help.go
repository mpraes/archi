package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func attachStyledHelp(root *cobra.Command) {
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if cmd.Name() != root.Name() {
			fmt.Fprintln(cmd.OutOrStdout(), renderSubcommandHelp(cmd))
			return
		}
		fmt.Fprintln(cmd.OutOrStdout(), renderRootHelp(cmd))
	})
}

func renderRootHelp(cmd *cobra.Command) string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	section := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		MarginTop(1)
	key := lipgloss.NewStyle().
		Foreground(lipgloss.Color("229")).
		Bold(true)
	value := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))
	bullet := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Render("•")

	var b strings.Builder
	b.WriteString(title.Render("ARCHI"))
	b.WriteString(" ")
	b.WriteString(subtitle.Render("diagnóstico arquitetural local"))
	b.WriteString("\n")
	b.WriteString(subtitle.Render("Mapeie acoplamentos, hotspots e rigidez com visualização interativa."))
	b.WriteString("\n")

	b.WriteString(section.Render("Uso rápido"))
	b.WriteString("\n")
	b.WriteString(value.Render("  archi [caminho_do_projeto] [flags]"))
	b.WriteString("\n")
	b.WriteString(value.Render("  archi [command]"))
	b.WriteString("\n")

	b.WriteString(section.Render("Comandos"))
	b.WriteString("\n")
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		b.WriteString("  ")
		b.WriteString(bullet)
		b.WriteString(" ")
		b.WriteString(key.Render(c.Name()))
		b.WriteString("  ")
		b.WriteString(value.Render(c.Short))
		b.WriteString("\n")
	}

	b.WriteString(section.Render("Flags"))
	b.WriteString("\n")
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		var name string
		if f.Shorthand != "" {
			name = fmt.Sprintf("-%s, --%s", f.Shorthand, f.Name)
		} else {
			name = fmt.Sprintf("    --%s", f.Name)
		}
		b.WriteString("  ")
		b.WriteString(key.Render(name))
		b.WriteString("\n      ")
		b.WriteString(value.Render(f.Usage))
		if f.DefValue != "" && f.DefValue != "false" {
			b.WriteString(value.Render(fmt.Sprintf(" (default: %s)", f.DefValue)))
		}
		b.WriteString("\n")
	})

	b.WriteString(section.Render("Primeiros passos"))
	b.WriteString("\n")
	b.WriteString(value.Render("  1) Analisar sem abrir navegador automático:"))
	b.WriteString("\n")
	b.WriteString(value.Render("     archi . --no-browser"))
	b.WriteString("\n")
	b.WriteString(value.Render("  2) Exportar relatório para CI/pipeline:"))
	b.WriteString("\n")
	b.WriteString(value.Render("     archi export --format json"))
	b.WriteString("\n")
	b.WriteString(value.Render("  3) Validar gate arquitetural no CI:"))
	b.WriteString("\n")
	b.WriteString(value.Render("     archi check --max-distance 0.7"))
	b.WriteString("\n")
	b.WriteString("\n")
	b.WriteString(subtitle.Render(`Dica: use "archi [command] --help" para ajuda detalhada por comando.`))

	return b.String()
}

func renderSubcommandHelp(cmd *cobra.Command) string {
	var b strings.Builder
	b.WriteString(cmd.Short)
	if cmd.Long != "" {
		b.WriteString("\n\n")
		b.WriteString(cmd.Long)
	}
	b.WriteString("\n\nUsage:\n  ")
	b.WriteString(cmd.CommandPath())
	if cmd.Use != "" {
		parts := strings.Fields(cmd.Use)
		if len(parts) > 1 {
			b.WriteString(" ")
			b.WriteString(strings.Join(parts[1:], " "))
		}
	}
	b.WriteString("\n")

	b.WriteString("\nFlags:\n")
	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		renderFlagLine(&b, f)
	})
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		renderFlagLine(&b, f)
	})
	return b.String()
}

func renderFlagLine(b *strings.Builder, f *pflag.Flag) {
	if f.Shorthand != "" {
		b.WriteString(fmt.Sprintf("  -%s, --%s", f.Shorthand, f.Name))
	} else {
		b.WriteString(fmt.Sprintf("      --%s", f.Name))
	}
	if f.Value.Type() != "bool" {
		b.WriteString(" ")
		b.WriteString(f.Value.Type())
	}
	b.WriteString("\n      ")
	b.WriteString(f.Usage)
	if f.DefValue != "" && f.DefValue != "false" {
		b.WriteString(fmt.Sprintf(" (default: %s)", f.DefValue))
	}
	b.WriteString("\n")
}
