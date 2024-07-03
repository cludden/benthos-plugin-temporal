package connect

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"

	"github.com/redpanda-data/benthos/v4/public/service"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/filter/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func TestConnectWorkflowOutput_Basic(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	r, ctx := require.New(t), context.Background()

	srv, err := testsuite.StartDevServer(ctx, testsuite.DevServerOptions{})
	r.NoError(err)
	t.Cleanup(func() {
		r.NoError(srv.Stop())
	})

	c := srv.Client()
	t.Cleanup(c.Close)

	w := worker.New(c, "test", worker.Options{})
	w.RegisterWorkflowWithOptions(func(ctx workflow.Context, input map[string]any) (map[string]any, error) {
		return input, nil
	}, workflow.RegisterOptions{Name: "basic"})
	r.NoError(w.Start())
	t.Cleanup(w.Stop)

	builder := service.NewStreamBuilder()
	builder.SetLogger(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	producer, err := builder.AddProducerFunc()
	r.NoError(err)
	r.NoError(builder.AddOutputYAML(fmt.Sprintf(`
temporal_workflow:
  address: %s
  task_queue: test
  workflow_id: basic/${! uuid_v4() }
  workflow_type: basic
`, srv.FrontendHostPort())))
	stream, err := builder.Build()
	r.NoError(err)

	var g sync.WaitGroup
	g.Add(1)
	go func() {
		defer g.Done()
		r.NoError(stream.Run(ctx))
	}()

	r.NoError(producer(ctx, service.NewMessage([]byte(`{"foo":"bar"}`))))
	r.NoError(stream.Stop(ctx))
	g.Wait()

	workflows, err := c.WorkflowService().ListClosedWorkflowExecutions(ctx, &workflowservice.ListClosedWorkflowExecutionsRequest{
		Namespace: "default",
		Filters: &workflowservice.ListClosedWorkflowExecutionsRequest_TypeFilter{
			TypeFilter: &filter.WorkflowTypeFilter{
				Name: "basic",
			},
		},
	})
	r.NoError(err)
	r.NotNil(workflows)
	r.Len(workflows.Executions, 1)
	exec := workflows.GetExecutions()[0].GetExecution()

	result := make(map[string]any)
	r.NoError(c.GetWorkflow(ctx, exec.GetWorkflowId(), exec.GetRunId()).Get(ctx, &result))
	r.Equal(map[string]any{"foo": "bar"}, result)
}
