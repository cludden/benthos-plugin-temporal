# benthos-plugin-temporal

A [Temporal](https://temporal.io) integration for benthos/[redpanda-connect](https://docs.redpanda.com/redpanda-connect/about/)/[bento](https://warpstreamlabs.github.io/bento/).

Easily execute Temporal workflows from:

>  `amqp`, `aws_kinesis`, `aws_s3`, `aws_sqs`, `azure_blob_storage`,
>  `azure_cosmosdb`, `azure_queue_storage`, `azure_table_storage`, `beanstalkd`, 
>  `cassandra`, `discord`, `gcp_bigquery_select`, `gcp_cloud_storage`, `gcp_pubsub`, 
>  `hdfs`, `kafka` , `mongodb`, `mqtt`, `nats`, `nats_jetstream`, `nsq`, `pulsar`, 
>  `redis_list`, `redis_pubsub`, `redis_scan`, `redis_streams`, `sftp`, `sql`,
>  `twitter`, `webhooks`, and more!

## Getting Started

Launch a webhook server for executing dynamic workflows:

1. Import this plugin
  - [Redpanda Connect](https://docs.redpanda.com/redpanda-connect/about/)
    ```go
    // in main.go
    package main

    import (
        _ "github.com/cludden/benthos-plugin-temporal/pkg/connect/all"
        "github.com/redpanda-data/benthos/v4/public/service"
        _ "github.com/redpanda-data/connect/public/bundle/free/v4"
    )

    func main() {
        service.RunCLI(context.Background())
    }
    ```
  - [Bento](https://warpstreamlabs.github.io/bento/)
    ```go
    // in main.go
    package main

    import (
        _ "github.com/cludden/benthos-plugin-temporal/pkg/bento/all"
        _ "github.com/warpstreamlabs/bento/public/components/all"
	      "github.com/warpstreamlabs/bento/public/service"
    )

    func main() {
        service.RunCLI(context.Background())
    }
    ```
2. Use plugin in stream configuration
```yaml
# in config.yml
input:
  http_server:
    address: 0.0.0.0:8080
    path: /temporal

pipeline:
  processors:
    # extract task_queue, workflow_type, and workflow_id from http headers, query parameters, or json payload
    - mapping: |
        root = this.without("@task_queue", "@workflow_type", "@workflow_id")
        root."@metadata" = @
        meta task_queue = @."task_queue".or(this."@task_queue").or("example")
        meta workflow_id = @."workflow_id".or(this."@workflow_id").or("example/%s".format(uuid_v4()))
        meta workflow_type = @."workflow_type".or(this."@workflow_type").or("example")

output:
  temporal_workflow:
    address: localhost:7233
    task_queue: '${! @.task_queue }'
    workflow_id: '${! @.workflow_id }'
    workflow_type: '${! @.workflow_type }'
```
3. Start Temporal server
```shell
temporal server start-dev
```
4. In a separate terminal, run the application
```shell
go run main.go -c config.yml
```
5. In a separate terminal, trigger a workflow
```shell
curl http://localhost:8080/temporal \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"@task_queue":"example","@workflow_type":"example","foo":"bar"}' 
```

## Examples

See the [example](./example/) directory for complete examples.

## Resources

### Processors

#### verify_hmac_sha256

securely verifies an hmac_sha256 signature without leaking timing information

##### Fields

- secret (`<InterpolatedString>`) - hmac secret
- signature (`<InterpolatedString>`) - signature to verify (excluding any prefix/suffix)
- string_to_sign (`<Mapping>`) - string to sign

##### Example

**GitHub Webhook:**

```yaml
input:
  http_server:
    path: /post

pipeline:
  processors:
    - verify_hmac_sha256:
        secret: '${! env("WEBHOOK_SECRET") }'
        signature: '${! @."X-Hub-Signature-256".trim_prefix("sha256=") }'
        string_to_sign: root = content()

output:
  reject_errored:
    sync_response: {}
```

**Slack Request:**

```yaml
input:
  http_server:
    path: /post

pipeline:
  processors:
    - verify_hmac_sha256:
        secret: '${! env("WEBHOOK_SECRET") }'
        signature: '${! @."X-Slack-Signature".trim_prefix("v0=") }'
        string_to_sign: root = "v0:%s:%s".format(@."X-Slack-Request-Timestamp", content())

output:
  reject_errored:
    sync_response: {}
```

### Outputs

#### temporal_workflow

executes a Temporal workflow for each message as input

##### Fields

- address `<string>` - temporal cluster address
- codec_auth `[string]` - codec endpoint authorization header
- codec_endpoint `[string]` - remote codec server endpoint
- detach `[InterpolatedString]` - boolean indicating whether the output should wait for workflow completion before acknowleding a message
- namespace `[string]` - temporal namespace name
- search_attributes `[Mapping]` - bloblang mapping defining workflow search attributes
- task_queue `<InterpolatedString>` - temporal worker task queue name
- tls.ca_data `[string]` - pem-encoded ca data
- tls.ca_file `[string]` - path to pem-encoded ca certificate
- tls.cert_data `[string]` - pem-encoded client certificate data
- tls.cert_file `[string]` - path to pem-encoded client certificate
- tls.disable_host_verification `[bool]` - disables tls host verification
- tls.key_data `[string]` - pem-encoded client private key
- tls.key_file `[string]` - path to pem-encoded client private key
- tls.server_name `[string]` - overrides target tls server name
- workflow_id `<InterpolatedString>` - temporal workflow id
- workflow_type `<InterpolatedString>` - temporal workflow type

##### Example

```yaml
output:
  temporal_workflow:
    address: localhost:7233
    task_queue: example
    workflow_id: test/${! uuid_v4() }
    workflow_type: ${! @.workflow_type.or(this."@workflow_type").or("test") }
```

## License
Licensed under the [MIT License](LICENSE.md)  
Copyright (c) 2024 Chris Ludden
