package verifyhmacsha256processor

import (
	"fmt"

	"github.com/cludden/benthos-plugin-temporal/pkg/bento"
	"github.com/cludden/benthos-plugin-temporal/pkg/plugin"
	"github.com/warpstreamlabs/bento/public/service"
)

func init() {
	if err := service.RegisterProcessor(plugin.VerifyHmacSha256ProcessorType, plugin.NewVerifyHmacSha256ProcessorConfig(service.NewConfigSpec(), bento.DefaultFieldProvider), func(conf *service.ParsedConfig, mgr *service.Resources) (service.Processor, error) {
		return plugin.NewVerifyHmacSha256Processor(conf, mgr, bento.MessageBatch)
	}); err != nil {
		panic(fmt.Errorf("error registering %s processor: %w", plugin.VerifyHmacSha256ProcessorType, err))
	}
}
