package plugin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"

	"github.com/cludden/protoc-gen-go-temporal/pkg/scheme"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	WorkflowOutputType = "temporal_workflow"
)

type (
	WorkflowOutput[
		InterpolatedString interface {
			TryString(Message) (string, error)
		},
		Mapping BloblangMapping,
		Message interface {
			AsBytes() ([]byte, error)
			AsStructured() (any, error)
			BloblangQuery(Mapping) (Message, error)
		},
	] struct {
		client                 client.Client
		clientOpts             client.Options
		dc                     converter.DataConverter
		detach                 InterpolatedString
		mapping                Mapping
		mappingExists          bool
		inputMessageType       InterpolatedString
		inputMessageTypeExists bool
		scheme                 *scheme.Scheme
		searchAttributes       Mapping
		searchAttributesExists bool
		taskQueue              InterpolatedString
		workflowID             InterpolatedString
		workflowType           InterpolatedString
	}

	WorkflowOutputOptions[
		InterpolatedString interface {
			TryString(Message) (string, error)
		},
		Mapping BloblangMapping,
		Message interface {
			AsBytes() ([]byte, error)
			AsStructured() (any, error)
			BloblangQuery(Mapping) (Message, error)
		},
	] func(*WorkflowOutput[InterpolatedString, Mapping, Message]) error
)

func NewWorkflowOutputConfig[
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
	return conf.Summary("Executes a Temporal workflow for each message as input.").
		Fields(
			fields.NewStringField("address").
				Description("Temporal cluster address"),
			fields.NewStringField("codec_auth").
				Description("Authorization header for requests to Codec Server").
				Optional(),
			fields.NewStringField("codec_endpoint").
				Description("Endpoint for remote Codec Server").
				Optional(),
			fields.NewInterpolatedStringEnumField("detach", "true", "false").
				Description("Starts the workflow execution without waiting for the result").
				Default("false"),
			fields.NewInterpolatedStringField("input_proto_message_name").
				Description("Full name of input proto message").
				Optional(),
			fields.NewBloblangField("mapping").
				Description("Input mapping").
				Optional(),
			fields.NewIntField("max_in_flight").
				Description("Maximum number of pending workflow executions").
				Default(1),
			fields.NewStringField("namespace").
				Description("Temporal namespace name").
				Default("default"),
			fields.NewBloblangField("search_attributes").
				Description("Search attributes mapping").
				Optional(),
			fields.NewInterpolatedStringField("task_queue").
				Description("Worker task queue name"),
			fields.NewObjectField("tls",
				fields.NewStringField("ca_file").
					Description("Path to ca file").
					Optional(),
				fields.NewStringField("ca_data").
					Description("PEM-encoded ca data").
					Optional(),
				fields.NewStringField("cert_file").
					Description("Path to certificate file").
					Optional(),
				fields.NewStringField("cert_data").
					Description("PEM-encoded certificate data").
					Optional(),
				fields.NewBoolField("disable_host_verification").
					Description("Disable TLS host verification").
					Optional(),
				fields.NewStringField("key_file").
					Description("Path to private key").
					Optional(),
				fields.NewStringField("key_data").
					Description("PEM-encoded private key data").
					Optional(),
				fields.NewStringField("server_name").
					Description("Override target TLS server name").
					Optional(),
			).
				Description("Optional TLS configuration").
				Optional(),
			fields.NewInterpolatedStringField("workflow_id").
				Description("Workflow ID"),
			fields.NewInterpolatedStringField("workflow_type").
				Description("Workflow type name"),
		)
}

func NewWorkflowOutput[
	InterpolatedString interface {
		TryString(Message) (string, error)
	},
	Mapping BloblangMapping,
	Message interface {
		AsBytes() ([]byte, error)
		AsStructured() (any, error)
		BloblangQuery(Mapping) (Message, error)
	},
	ParsedConfig interface {
		Contains(...string) bool
		FieldBloblang(...string) (Mapping, error)
		FieldBool(...string) (bool, error)
		FieldInt(...string) (int, error)
		FieldInterpolatedString(...string) (InterpolatedString, error)
		FieldString(...string) (string, error)
	},
	Resources any,
](conf ParsedConfig, mgr Resources, opts ...WorkflowOutputOptions[InterpolatedString, Mapping, Message]) (o *WorkflowOutput[InterpolatedString, Mapping, Message], maxInFlight int, err error) {
	o = &WorkflowOutput[InterpolatedString, Mapping, Message]{}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, 0, err
		}
	}
	if o.clientOpts.HostPort, err = conf.FieldString("address"); err != nil {
		return nil, 0, err
	}
	if o.dc == nil {
		o.dc = converter.GetDefaultDataConverter()
	}
	if conf.Contains("codec_endpoint") {
		var codecOpts converter.RemotePayloadCodecOptions
		if codecOpts.Endpoint, err = conf.FieldString("codec_endpoint"); err != nil {
			return nil, 0, err
		}
		if conf.Contains("codec_auth") {
			codecAuth, err := conf.FieldString("codec_auth")
			if err != nil {
				return nil, 0, err
			}
			codecOpts.ModifyRequest = func(r *http.Request) error {
				r.Header.Set("Authorization", codecAuth)
				return nil
			}
		}
		o.dc = converter.NewCodecDataConverter(o.dc, converter.NewRemotePayloadCodec(codecOpts))
	}
	if o.detach, err = conf.FieldInterpolatedString("detach"); err != nil {
		return nil, 0, err
	}
	if conf.Contains("input_proto_message_name") {
		o.inputMessageTypeExists = true
		if o.inputMessageType, err = conf.FieldInterpolatedString("input_proto_message_name"); err != nil {
			return nil, 0, err
		}
	}
	if conf.Contains("mapping") {
		o.mappingExists = true
		if o.mapping, err = conf.FieldBloblang("mapping"); err != nil {
			return nil, 0, err
		}
	}
	if maxInFlight, err = conf.FieldInt("max_in_flight"); err != nil {
		return nil, 0, err
	}
	if o.clientOpts.Namespace, err = conf.FieldString("namespace"); err != nil {
		return nil, 0, err
	}
	if conf.Contains("search_attributes") {
		o.searchAttributesExists = true
		if o.searchAttributes, err = conf.FieldBloblang("search_attributes"); err != nil {
			return nil, 0, err
		}
	}
	if o.taskQueue, err = conf.FieldInterpolatedString("task_queue"); err != nil {
		return nil, 0, err
	}
	if o.clientOpts.ConnectionOptions.TLS, err = parseTLS[InterpolatedString, Mapping, Message, ParsedConfig](conf); err != nil {
		return nil, 0, err
	}
	if o.workflowID, err = conf.FieldInterpolatedString("workflow_id"); err != nil {
		return nil, 0, err
	}
	if o.workflowType, err = conf.FieldInterpolatedString("workflow_type"); err != nil {
		return nil, 0, err
	}
	return o, maxInFlight, nil
}

func (o *WorkflowOutput[InterpolatedString, Mapping, Message]) Close(ctx context.Context) error {
	o.client.Close()
	return nil
}

func (o *WorkflowOutput[InterpolatedString, Mapping, Message]) Connect(ctx context.Context) (err error) {
	if o.client, err = client.Dial(o.clientOpts); err != nil {
		return fmt.Errorf("error connecting to Temporal: %w", err)
	}
	return nil
}

func (o *WorkflowOutput[InterpolatedString, Mapping, Message]) Write(ctx context.Context, msg Message) (err error) {
	var opts client.StartWorkflowOptions
	if opts.ID, err = o.workflowID.TryString(msg); err != nil {
		return fmt.Errorf("error evaluating workflow_id: %w", err)
	}
	if opts.TaskQueue, err = o.taskQueue.TryString(msg); err != nil {
		return fmt.Errorf("error evaluating task_queue: %w", err)
	}
	workflowType, err := o.workflowType.TryString(msg)
	if err != nil {
		return fmt.Errorf("error evaluating workflow_type: %w", err)
	}
	if o.mappingExists {
		if msg, err = msg.BloblangQuery(o.mapping); err != nil {
			return fmt.Errorf("error applying output mapping: %w", err)
		}
	}
	if o.searchAttributesExists {
		searchAttributes, err := o.searchAttributes.Query(msg)
		if err != nil {
			return fmt.Errorf("error evaluating search_attributes: %w", err)
		}
		sa, ok := searchAttributes.(map[string]any)
		if !ok {
			return fmt.Errorf("expected search_attributes to return an object, got: %T", searchAttributes)
		}
		opts.SearchAttributes = sa
	}

	var run client.WorkflowRun
	var empty Message
	if !reflect.DeepEqual(msg, empty) {
		var arg any
		if o.scheme != nil && o.inputMessageTypeExists {
			messageType, err := o.inputMessageType.TryString(msg)
			if err != nil {
				return fmt.Errorf("error evaluating input_proto_message_name: %w", err)
			}
			pb, err := o.scheme.New(messageType)
			if err != nil {
				return fmt.Errorf("error initializing new %s value: %w", messageType, err)
			}
			b, err := msg.AsBytes()
			if err != nil {
				return fmt.Errorf("error serializing message bytes: %w", err)
			}
			if err = protojson.Unmarshal(b, pb); err != nil {
				return fmt.Errorf("error unmarshalling message proto: %w", err)
			}
			arg = pb
		} else if arg, err = msg.AsStructured(); err != nil {
			return fmt.Errorf("error evaluating message as structured: %w", err)
		}

		run, err = o.client.ExecuteWorkflow(ctx, opts, workflowType, arg)
	} else {
		run, err = o.client.ExecuteWorkflow(ctx, opts, workflowType)
	}
	if err != nil {
		return fmt.Errorf("error executing workflow: %w", err)
	}
	if detach, _ := o.detach.TryString(msg); detach == "true" {
		return nil
	}
	return run.Get(ctx, nil)
}

func parseTLS[
	InterpolatedString interface {
		TryString(Message) (string, error)
	},
	Mapping BloblangMapping,
	Message interface {
		AsBytes() ([]byte, error)
		AsStructured() (any, error)
		BloblangQuery(Mapping) (Message, error)
	},
	ParsedConfig interface {
		Contains(...string) bool
		FieldBloblang(...string) (Mapping, error)
		FieldBool(...string) (bool, error)
		FieldInt(...string) (int, error)
		FieldInterpolatedString(...string) (InterpolatedString, error)
		FieldString(...string) (string, error)
	},
](conf ParsedConfig) (cfg *tls.Config, err error) {
	cfg = &tls.Config{}

	var caBytes []byte
	if caFile, _ := conf.FieldString("tls", "ca_file"); caFile != "" {
		if conf.Contains("tls", "ca_data") {
			return nil, errors.New("cannot specify both ca_data and ca_file")
		}
		if caBytes, err = os.ReadFile(caFile); err != nil {
			return nil, err
		}
	} else if caData, _ := conf.FieldString("tls", "ca_data"); caData != "" {
		caBytes = []byte(caData)
	}
	if len(caBytes) > 0 {
		cfg.RootCAs = x509.NewCertPool()
		if !cfg.RootCAs.AppendCertsFromPEM(caBytes) {
			return nil, errors.New("invalid CA cert data")
		}
	}

	var clientCert tls.Certificate
	var hasClientCert bool
	if conf.Contains("tls", "cert_file") && conf.Contains("tls", "key_file") {
		certFile, _ := conf.FieldString("tls", "cert_file")
		keyFile, _ := conf.FieldString("tls", "key_file")
		clientCert, err = tls.LoadX509KeyPair(certFile, keyFile)
		hasClientCert = true
	} else if conf.Contains("tls", "cert_data") && conf.Contains("tls", "key_data") {
		certData, _ := conf.FieldString("tls", "cert_data")
		keyData, _ := conf.FieldString("tls", "key_data")
		clientCert, err = tls.X509KeyPair([]byte(certData), []byte(keyData))
		hasClientCert = true
	}
	if err != nil {
		return nil, fmt.Errorf("error loading client certificate: %w", err)
	}
	if hasClientCert {
		cfg.Certificates = append(cfg.Certificates, clientCert)
		hasClientCert = true
	}
	if conf.Contains("tls", "disable_host_verification") {
		if cfg.InsecureSkipVerify, err = conf.FieldBool("tls", "disable_host_verification"); err != nil {
			return nil, err
		}
	}
	if conf.Contains("tls", "server_name") {
		if cfg.ServerName, err = conf.FieldString("tls", "server_name"); err != nil {
			return nil, err
		}
	}
	if len(cfg.Certificates) > 0 || cfg.InsecureSkipVerify || cfg.RootCAs != nil || cfg.ServerName != "" {
		return cfg, nil
	}
	return nil, nil
}
