package workflowoutput

import (
	"fmt"

	"github.com/cludden/benthos-plugin-temporal/pkg/connect"
	"github.com/cludden/benthos-plugin-temporal/pkg/plugin"
	"github.com/redpanda-data/benthos/v4/public/service"
)

func init() {
	if err := service.RegisterOutput(plugin.WorkflowOutputType, plugin.NewWorkflowOutputConfig(service.NewConfigSpec(), connect.DefaultFieldProvider), func(conf *service.ParsedConfig, mgr *service.Resources) (service.Output, int, error) {
		return plugin.NewWorkflowOutput(conf, mgr)
	}); err != nil {
		panic(fmt.Errorf("error registering %s output: %w", plugin.WorkflowOutputType, err))
	}
}
