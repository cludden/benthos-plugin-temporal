package plugin

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const (
	VerifyHmacSha256ProcessorType = "verify_hmac_sha256"
)

type (
	VerifyHmacSha256Processor[
		InterpolatedString interface {
			TryString(Message) (string, error)
		},
		Mapping BloblangMapping,
		Message interface {
			AsBytes() ([]byte, error)
			AsStructured() (any, error)
			BloblangQuery(Mapping) (Message, error)
		},
		MessageBatch any,
	] struct {
		secret       InterpolatedString
		signature    InterpolatedString
		stringToSign Mapping
		toBatch      func([]Message) MessageBatch
	}
)

func NewVerifyHmacSha256ProcessorConfig[
	Field interface {
		Default(any) Field
		Description(string) Field
		Optional() Field
	},
	ConfigSpec interface {
		Summary(string) ConfigSpec
		Fields(...Field) ConfigSpec
	},
	FieldProvider interface {
		NewBoolField(string) Field
		NewBloblangField(string) Field
		NewIntField(string) Field
		NewStringField(string) Field
		NewInterpolatedStringEnumField(string, ...string) Field
		NewInterpolatedStringField(string) Field
		NewObjectField(string, ...Field) Field
	},
](conf ConfigSpec, fields FieldProvider) ConfigSpec {
	return conf.Summary("Securely verifies an hmac_sha256 signature without leaking timing information.").
		Fields(
			fields.NewInterpolatedStringField("secret").
				Description("hmac secret used to verify signature"),
			fields.NewInterpolatedStringField("signature").
				Description("expected signature"),
			fields.NewBloblangField("string_to_sign").
				Description("source message"),
		)
}

func NewVerifyHmacSha256Processor[
	InterpolatedString interface {
		TryString(Message) (string, error)
	},
	Mapping BloblangMapping,
	Message interface {
		AsBytes() ([]byte, error)
		AsStructured() (any, error)
		BloblangQuery(Mapping) (Message, error)
	},
	MessageBatch any,
	ParsedConfig interface {
		Contains(...string) bool
		FieldBloblang(...string) (Mapping, error)
		FieldBool(...string) (bool, error)
		FieldInt(...string) (int, error)
		FieldInterpolatedString(...string) (InterpolatedString, error)
		FieldString(...string) (string, error)
	},
	Resources any,
](conf ParsedConfig, mgr Resources, toBatch func([]Message) MessageBatch) (o *VerifyHmacSha256Processor[InterpolatedString, Mapping, Message, MessageBatch], err error) {
	secret, err := conf.FieldInterpolatedString("secret")
	if err != nil {
		return nil, err
	}
	signature, err := conf.FieldInterpolatedString("signature")
	if err != nil {
		return nil, err
	}
	stringToSign, err := conf.FieldBloblang("string_to_sign")
	if err != nil {
		return nil, err
	}
	return &VerifyHmacSha256Processor[InterpolatedString, Mapping, Message, MessageBatch]{
		secret:       secret,
		signature:    signature,
		stringToSign: stringToSign,
		toBatch:      toBatch,
	}, nil
}

func (p *VerifyHmacSha256Processor[InterpolatedString, Mapping, Message, MessageBatch]) Close(ctx context.Context) error {
	return nil
}

func (p *VerifyHmacSha256Processor[InterpolatedString, Mapping, Message, MessageBatch]) Process(ctx context.Context, msg Message) (result MessageBatch, err error) {
	secret, err := p.secret.TryString(msg)
	if err != nil {
		return result, fmt.Errorf("error evaluating secret: %w", err)
	}
	signature, err := p.signature.TryString(msg)
	if err != nil {
		return result, fmt.Errorf("error evaluating signature: %w", err)
	}
	stringToSign, err := msg.BloblangQuery(p.stringToSign)
	if err != nil {
		return result, fmt.Errorf("error evaluating string_to_sign: %w", err)
	}
	b, err := stringToSign.AsBytes()
	if err != nil {
		return result, fmt.Errorf("invalid result from string_to_sign: %w", err)
	}
	h := hmac.New(sha256.New, []byte(secret))
	if _, err := h.Write(b); err != nil {
		return result, fmt.Errorf("error hashing string_to_sign: %w", err)
	}
	if !hmac.Equal([]byte(signature), []byte(hex.EncodeToString(h.Sum(nil)))) {
		return result, fmt.Errorf("invalid signature")
	}
	return p.toBatch([]Message{msg}), nil
}
