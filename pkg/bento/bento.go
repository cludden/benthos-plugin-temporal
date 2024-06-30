package temporal

import (
	"fmt"

	plugin "github.com/cludden/benthos-plugin-temporal"
	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	if err := service.RegisterOutput(plugin.WorkflowOutputType, WorkflowOutputConfig, func(conf *service.ParsedConfig, mgr *service.Resources) (service.Output, int, error) {
		return plugin.NewWorkflowOutput(conf, mgr)
	}); err != nil {
		panic(fmt.Errorf("error registering %s output: %w", plugin.WorkflowOutputType, err))
	}
}

var (
	WorkflowOutputConfig = plugin.WorkflowOutputConfig(service.NewConfigSpec(), &fieldProvider{})
)

type fieldProvider struct{}

func (fp *fieldProvider) NewBloblangField(name string) *service.ConfigField {
	return service.NewBloblangField(name)
}

func (fp *fieldProvider) NewBoolField(name string) *service.ConfigField {
	return service.NewBoolField(name)
}

func (fp *fieldProvider) NewInterpolatedStringEnumField(name string, values ...string) *service.ConfigField {
	return service.NewInterpolatedStringEnumField(name, values...)
}

func (fp *fieldProvider) NewInterpolatedStringField(name string) *service.ConfigField {
	return service.NewInterpolatedStringField(name)
}

func (fp *fieldProvider) NewIntField(name string) *service.ConfigField {
	return service.NewIntField(name)
}

func (fp *fieldProvider) NewObjectField(name string, fields ...*service.ConfigField) *service.ConfigField {
	return service.NewObjectField(name, fields...)
}

func (fp *fieldProvider) NewStringField(name string) *service.ConfigField {
	return service.NewStringField(name)
}
