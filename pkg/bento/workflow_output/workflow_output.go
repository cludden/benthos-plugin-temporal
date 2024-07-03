package workflowoutput

import (
	"fmt"

	"github.com/cludden/benthos-plugin-temporal/pkg/bento"
	"github.com/cludden/benthos-plugin-temporal/pkg/plugin"
	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	if err := service.RegisterOutput(plugin.WorkflowOutputType, plugin.NewWorkflowOutputConfig(service.NewConfigSpec(), bento.DefaultFieldProvider), func(conf *service.ParsedConfig, mgr *service.Resources) (service.Output, int, error) {
		return plugin.NewWorkflowOutput(conf, mgr)
	}); err != nil {
		panic(fmt.Errorf("error registering %s output: %w", plugin.WorkflowOutputType, err))
	}
}
