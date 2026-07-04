package cli

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/mpraes/archi/internal/ai"
	"github.com/mpraes/archi/internal/analyzer"
	"github.com/mpraes/archi/internal/model"
	"github.com/mpraes/archi/internal/parser"
	"github.com/mpraes/archi/internal/server"
	"github.com/mpraes/archi/internal/ui"
)

func runRoot(cmd *cobra.Command, args []string, g *GlobalFlags) error {
	path := g.pathArg(args)
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	summary, warnings := scan(path, g)
	for _, w := range warnings {
		logger.Warn("parser", "warning", w)
	}

	aiCfg := aiEnabled(g)

	srv := server.New(summary, aiCfg, logger)
	addr := fmt.Sprintf("127.0.0.1:%d", g.Port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	addr = ln.Addr().String()

	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Run(ctx, ln) }()

	if !g.NoBrowser {
		url := "http://" + addr
		if err := ui.OpenBrowser(url); err != nil {
			logger.Warn("browser", "open", err)
		} else {
			fmt.Fprintln(os.Stderr, "Archi aberto em", url)
		}
	} else {
		fmt.Fprintln(os.Stderr, "Archi servindo em http://"+addr)
	}

	fmt.Fprintln(os.Stderr, "Ctrl-C para encerrar.")
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// scan parses and analyzes the project, returning the metrics summary plus
// parser warnings (never fatal).
func scan(path string, g *GlobalFlags) (model.Summary, []string) {
	spin := ui.NewSpinner("Analisando estrutura...")
	spin.Start()
	defer spin.Stop()

	prog := parser.Parse(path, parser.Options{
		Lang:    g.Lang,
		Exclude: g.Exclude,
	})
	s := analyzer.Analyze(prog.Program)
	return s, prog.Warnings
}

func aiEnabled(g *GlobalFlags) *ai.Config {
	key := g.APIKey
	if key == "" {
		key = os.Getenv("GEMINI_API_KEY")
	}
	if !g.AIPresent && key == "" {
		return nil
	}
	return &ai.Config{APIKey: key}
}