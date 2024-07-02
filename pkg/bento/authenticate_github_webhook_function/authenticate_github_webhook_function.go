package authenticategithubwebhookfunction

import (
	"fmt"

	"github.com/cludden/benthos-plugin-temporal/pkg/bento"
	"github.com/cludden/benthos-plugin-temporal/pkg/plugin"
	"github.com/warpstreamlabs/bento/public/bloblang"
)

func init() {
	if err := bloblang.RegisterFunctionV2(plugin.AuthenticateGithubWebhookFunctionType, plugin.NewAuthenticateGithubWebhookFunctionConfig(bloblang.NewPluginSpec(), &bento.ParamProvider{}), func(args *bloblang.ParsedParams) (bloblang.Function, error) {
		fn, err := plugin.AuthenticateGithubWebhook(args)
		if err != nil {
			return nil, err
		}
		return bloblang.Function(fn), nil
	}); err != nil {
		panic(fmt.Errorf("error registering %s bloblang function: %w", plugin.AuthenticateGithubWebhookFunctionType, err))
	}
}
