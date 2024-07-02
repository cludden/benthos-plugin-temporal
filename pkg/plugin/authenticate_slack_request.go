package plugin

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const (
	AuthenticateSlackRequestFunctionType = "authenticate_slack_request"
)

func NewAuthenticateSlackRequestConfig[
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
		Description("authenticate a slack request").
		Param(
			params.NewStringParam("payload").
				Description("webhook payload"),
		).
		Param(
			params.NewStringParam("timestamp").
				Description("X-Slack-Request-Timestamp header value"),
		).
		Param(
			params.NewStringParam("secret").
				Description("webhook secret"),
		).
		Param(
			params.NewStringParam("signature").
				Description("webhook signature"),
		).
		Param(
			params.NewStringParam("grace_period").
				Description("maximum clock skew").
				Default("5s").
				Optional(),
		)
}

func AuthenticateSlackRequest[
	Function BloblangFunction,
	ParsedParams interface {
		GetString(string) (string, error)
	},
](params ParsedParams) (Function, error) {
	payload, err := params.GetString("payload")
	if err != nil {
		return nil, err
	}
	timestamp, err := params.GetString("timestamp")
	if err != nil {
		return nil, err
	}
	sig, err := params.GetString("signature")
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(sig, "v0=") {
		return nil, errors.New("signature should start with v0=")
	}
	secret, err := params.GetString("secret")
	if err != nil {
		return nil, err
	}
	return func() (any, error) {
		h := hmac.New(sha256.New, []byte(secret))
		if _, err := h.Write([]byte(fmt.Sprintf("v0:%s:%s", timestamp, payload))); err != nil {
			return nil, err
		}
		signature := fmt.Sprintf("v0=%s", hex.EncodeToString(h.Sum(nil)))
		if !hmac.Equal([]byte(sig), []byte(signature)) {
			return nil, errors.New("invalid signature")
		}
		return true, nil
	}, nil
}
