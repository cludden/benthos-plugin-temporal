package plugin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const (
	AuthenticateGithubWebhookFunctionType = "authenticate_github_webhook"
)

func NewAuthenticateGithubWebhookFunctionConfig[
	ParamDefinition interface {
		Default(any) ParamDefinition
		Description(string) ParamDefinition
		Optional() ParamDefinition
	},
	ParamProvider interface {
		NewStringParam(string) ParamDefinition
	},
	PluginSpec interface {
		Description(string) PluginSpec
		Param(ParamDefinition) PluginSpec
	},
](conf PluginSpec, params ParamProvider) PluginSpec {
	return conf.
		Description("authenticate a github webhook message").
		Param(
			params.NewStringParam("payload").
				Description("webhook payload"),
		).
		Param(
			params.NewStringParam("secret").
				Description("webhook secret"),
		).
		Param(
			params.NewStringParam("signature").
				Description("webhook signature"),
		)
}

func AuthenticateGithubWebhook[
	Function BloblangFunction,
	ParsedParams interface {
		GetString(string) (string, error)
	},
](params ParsedParams) (Function, error) {
	payload, err := params.GetString("payload")
	if err != nil {
		return nil, err
	}
	sig, err := params.GetString("signature")
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(sig, "sha256=") {
		return nil, errors.New("signature should start with sha256=")
	}
	sig = strings.TrimPrefix(sig, "sha256=")
	secret, err := params.GetString("secret")
	if err != nil {
		return nil, err
	}
	return func() (any, error) {
		h := hmac.New(sha256.New, []byte(secret))
		if _, err := h.Write([]byte(payload)); err != nil {
			return nil, err
		}
		signature := hex.EncodeToString(h.Sum(nil))
		if !hmac.Equal([]byte(sig), []byte(signature)) {
			return nil, errors.New("invalid signature")
		}
		return true, nil
	}, nil
}
