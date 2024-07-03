package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"syscall"
	"time"

	_ "github.com/cludden/benthos-plugin-temporal/pkg/connect/all"
	"github.com/oklog/run"
	"github.com/redpanda-data/benthos/v4/public/service"
	_ "github.com/redpanda-data/connect/public/bundle/free/v4"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	if err := exec(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func exec(ctx context.Context) error {
	// initialize temporal client
	c, err := client.Dial(client.Options{})
	if err != nil {
		return fmt.Errorf("error connecting to Temporal: %w", err)
	}
	defer c.Close()

	// initialize run group
	var g run.Group
	g.Add(run.SignalHandler(ctx, syscall.SIGINT, syscall.SIGTERM))

	// initialize and register temporal worker
	w := worker.New(c, "example", worker.Options{})
	w.RegisterWorkflowWithOptions(func(ctx workflow.Context, input map[string]any) (map[string]any, error) {
		var d time.Duration
		if err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
			return time.Millisecond * time.Duration(rand.Intn(500)+50)
		}).Get(&d); err != nil {
			return nil, err
		}
		return input, workflow.Sleep(ctx, d)
	}, workflow.RegisterOptions{Name: "example"})
	g.Add(
		func() error {
			return w.Run(nil)
		},
		func(error) {
			w.Stop()
		},
	)

	// initialize and register connect cli
	cliCtx, cliCancel := context.WithCancel(ctx)
	g.Add(
		func() error {
			service.RunCLI(cliCtx)
			return nil
		},
		func(error) {
			cliCancel()
		},
	)

	return g.Run()
}
