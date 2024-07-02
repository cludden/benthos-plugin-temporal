# benthos-plugin-temporal

A [Temporal](https://temporal.io) integration for benthos/[redpanda-connect](https://docs.redpanda.com/redpanda-connect/about/)/[bento](https://warpstreamlabs.github.io/bento/).

## Getting Started

1. Import this plugin
  - [Redpanda Connect](https://docs.redpanda.com/redpanda-connect/about/)
    ```go
    // in main.go
    package main

    import (
        _ "github.com/cludden/benthos-plugin-temporal/pkg/connect/workflow_output"
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
        _ "github.com/cludden/benthos-plugin-temporal/pkg/bento/workflow_output"
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
    # extract task_queue, workflow_type, and workflow_id from json payload or query parameters
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

## Resources

### Bloblang Functions

#### authenticate_github_webhook

authenticate a github webhook message

##### Fields

- payload (`string`) - webhook payload
- secret (`string`) - webhook secret
- signature (`string`) - value of `X-Hub-Signature-256` header

##### Example

```yaml
input:
  http_server:
    path: /post
    sync_response:
      status: '${! @.status.or(200) }'

pipeline:
  processors:
    - switch:
        - check: '!authenticate_github_webhook(payload: content(), secret: env("GITHUB_SECRET"), signature: meta("X-Hub-Signature-256")).catch(false)'
          processors:
            - mapping: meta status = 401

output:
  sync_response: {}
  processors:
    - mapping: root = {}
```

---

#### authenticate_slack_request

authenticate a slack request

##### Fields

- payload (`<string>`) - request payload
- timestamp (`<string>`) - value of `X-Slack-Request-Timestamp` header
- secret (`<string>`) - signing secret
- signature (`<string>`) - value of `X-Slack-Signature` header
- grace_period (`[string]`) - duration string containing the maximum allowed clock skew (default: `5s`)

##### Example

```yaml
input:
  http_server:
    path: /post
    sync_response:
      status: '${! @.status.or(200) }'

pipeline:
  processors:
    - switch:
        - check: '!authenticate_slack_request(payload: content(), timestamp: meta("X-Slack-Request-Timestamp") , secret: env("GITHUB_SECRET"), signature: meta("X-Slack-Signature")).catch(false)'
          processors:
            - mapping: meta status = 401

output:
  sync_response: {}
  processors:
    - mapping: root = {}
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
