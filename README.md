# benthos-plugin-temporal

A [Temporal](https://temporal.io) integration for benthos/[redpanda-connect](https://docs.redpanda.com/redpanda-connect/about/)/[bento](https://warpstreamlabs.github.io/bento/).

## Getting Started

1. Import this plugin
  - [Redpanda Connect](https://docs.redpanda.com/redpanda-connect/about/)
    ```go
    // in main.go
    package main

    import (
        _ "github.com/cludden/benthos-plugin-temporal/pkg/connect"
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
        _ "github.com/cludden/benthos-plugin-temporal/pkg/bento"
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
    - mapping: |
        root = this.without("@task_queue", "@workflow_type", "@workflow_id")
        root."@metadata" = @
        meta task_queue = @."task_queue".or(this."@task_queue").or("example")
        meta workflow_id = @."workflow_id".or(this."@workflow_id").or("example/%s".format(uuid_v4()))
        meta workflow_type = @."workflow_type".or(this."@workflow_type").or("example")
    - log:
        message: '${! content().string() }'
        fields_mapping: |
          task_queue = @.task_queue
          workflow_id = @.workflow_id
          workflow_type = @.workflow_type

output:
  temporal_workflow:
    address: localhost:7233
    task_queue: '${! @.task_queue }'
    workflow_id: '${! @.workflow_id }'
    workflow_type: '${! @.workflow_type }'
```
3. Start Temporal server
```shell
temporal server start-dev \                    
  --dynamic-config-value "frontend.enableUpdateWorkflowExecution=true" \
  --dynamic-config-value "frontend.enableUpdateWorkflowExecutionAsyncAccepted=true"
```
4. In a separate terminal, run the application
```shell
go run main.go -c config.yml
```
5. In a separate terminal, trigger a workflow
```shell
curl -XPOST -H "Content-Type: application/json" -d '{"foo":"bar"}' "http://localhost:8080/temporal?task_queue=example&workflow_type=example&workflow_id=test/123"
```

## License
Licensed under the [MIT License](LICENSE.md)  
Copyright (c) 2024 Chris Ludden
