package connect

import (
	"github.com/redpanda-data/benthos/v4/public/service"
)

var (
	DefaultFieldProvider = &FieldProvider{}
)

type FieldProvider struct{}

func (fp *FieldProvider) NewBloblangField(name string) *service.ConfigField {
	return service.NewBloblangField(name)
}

func (fp *FieldProvider) NewBoolField(name string) *service.ConfigField {
	return service.NewBoolField(name)
}

func (fp *FieldProvider) NewInterpolatedStringEnumField(name string, values ...string) *service.ConfigField {
	return service.NewInterpolatedStringEnumField(name, values...)
}

func (fp *FieldProvider) NewInterpolatedStringField(name string) *service.ConfigField {
	return service.NewInterpolatedStringField(name)
}

func (fp *FieldProvider) NewIntField(name string) *service.ConfigField {
	return service.NewIntField(name)
}

func (fp *FieldProvider) NewObjectField(name string, fields ...*service.ConfigField) *service.ConfigField {
	return service.NewObjectField(name, fields...)
}

func (fp *FieldProvider) NewStringField(name string) *service.ConfigField {
	return service.NewStringField(name)
}

func MessageBatch(msgs []*service.Message) service.MessageBatch {
	return service.MessageBatch(msgs)
}
