package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lossystyles/cli/internal/dashboard"
	"github.com/lossystyles/cli/internal/protocol"
	"github.com/lossystyles/cli/internal/server"
)

func main() {
	theme := flag.String("theme", "dark", "theme: dark, neon, retro, minimal, rainbow, pcr, eva")
	runID := flag.String("run-id", "", "run ID to listen for (creates socket at /tmp/lossystyles-{run-id}.sock)")
	demo := flag.Bool("demo", false, "run a demo with synthetic training data")
	flag.Parse()

	if *demo {
		runDemo(*theme)
		return
	}

	if *runID == "" {
		fmt.Fprintln(os.Stderr, "Usage: lossystyles --run-id <id> [--theme dark|neon|retro|minimal]")
		fmt.Fprintln(os.Stderr, "       lossystyles --demo [--theme dark|neon|retro|minimal]")
		os.Exit(1)
	}

	sockPath := server.SocketPath(*runID)
	srv, err := server.New(sockPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
	defer srv.Close()

	go srv.Accept()

	model := dashboard.New(srv.Messages, *theme)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runDemo simulates a training run with synthetic data.
func runDemo(themeName string) {
	msgs := make(chan protocol.Message, 256)

	// Feed synthetic metrics in a goroutine
	go func() {
		defer close(msgs)

		msgs <- protocol.Message{
			Type:    protocol.MsgInit,
			RunID:   "demo",
			Project: "gpt-micro",
			Config:  map[string]interface{}{"lr": 3e-4, "batch_size": 64},
		}

		totalSteps := 200
		baseLoss := 2.5
		lr := 3e-4

		for step := 0; step < totalSteps; step++ {
			// Simulate decreasing loss with noise
			progress := float64(step) / float64(totalSteps)
			loss := baseLoss * math.Exp(-3*progress) * (1 + 0.1*rand.Float64())
			accuracy := (1 - math.Exp(-4*progress)) * (0.95 + 0.05*rand.Float64())

			// LR warmup then cosine decay
			if step < 20 {
				lr = 3e-4 * float64(step) / 20.0
			} else {
				lr = 3e-4 * 0.5 * (1 + math.Cos(math.Pi*float64(step-20)/float64(totalSteps-20)))
			}

			// Gradient norm: spiky early, settles down
			gradNorm := 2.0*math.Exp(-2*progress) + 0.3*rand.Float64()

			msgs <- protocol.Message{
				Type:  protocol.MsgLog,
				RunID: "demo",
				Step:  step,
				Metrics: map[string]float64{
					"loss":      loss,
					"accuracy":  accuracy,
					"grad_norm": gradNorm,
					"lr":        lr,
				},
			}

			time.Sleep(80 * time.Millisecond)
		}

		msgs <- protocol.Message{Type: protocol.MsgFinish, RunID: "demo"}
	}()

	model := dashboard.New(msgs, themeName)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
