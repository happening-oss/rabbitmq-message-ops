<p align="center">
  <img src="assets/images/mascot.png" height="250" width="250"/>
</p>

<h1 align="center">
    rabbitmq-message-ops
</h1>

<p align="center">
    <a href="https://goreportcard.com/report/github.com/happening-oss/rabbitmq-message-ops">
        <img src="https://goreportcard.com/badge/github.com/happening-oss/rabbitmq-message-ops" alt="Go Report"/></a>
    <a href="https://github.com/happening-oss/rabbitmq-message-ops/actions/workflows/ci.yml">
        <img src="https://github.com/happening-oss/rabbitmq-message-ops/actions/workflows/ci.yml/badge.svg" alt="Build Status"/></a>
    <a href="https://github.com/happening-oss/rabbitmq-message-ops/releases">
        <img src="https://img.shields.io/github/v/release/happening-oss/rabbitmq-message-ops" alt="Latest Release"/></a>
    <a href="https://github.com/happening-oss/rabbitmq-message-ops/blob/master/LICENSE">
        <img src="https://img.shields.io/github/license/happening-oss/rabbitmq-message-ops" alt="MIT License"/></a>
</p>


`rabbitmq-message-ops` is an open-source command-line interface (CLI) tool for managing RabbitMQ messages.

## ✨ Features
- 👀 **View**: Retrieve messages from a specified queue.
- 🚚 **Move**: Move selected messages from one queue to another.
- 📑 **Copy**: Copy selected messages from one queue to another.
- 🧹 **Purge**: Remove messages from a queue based on a filter.
- 🔍 **Filtering**: Select specific messages with flexible filtering (**[expr-lang](https://expr-lang.org/docs/language-definition)**) based on message properties (see **[Filtering](#-filtering)** section).
- 📜 **Ordering**: Preserves the original order of messages (see **[Ordering](#-ordering)** section).

## ⚠️ Important Notice

The `rabbitmq-message-ops` tool is primarily designed for inspecting and fixing queue states.

> **⚠️ Caution**: The tool may behave unpredictably if the source queue is actively used by other consumers or producers.

To ensure the specified guarantees, only use the CLI tool when no other consumers or producers are interacting with the source queue.

## 🛠️ Prerequisites

Ensure you have the following installed:

- **[RabbitMQ](https://www.rabbitmq.com/download.html)**: up and running
  - Enabled RabbitMQ Management Plugin: `rabbitmq-plugins enable rabbitmq_management`
  - Ensure the HTTP API is accessible and your RabbitMQ user has sufficient permissions to interact with queues and messages.
- **[Go](https://golang.org/dl/)** v1.21 for building from source (optional)
- **[Taskfile](https://taskfile.dev/)** v3 for streamlined development (optional)

## ⚙️ Configuration

Before running any commands, it's recommended to set up the necessary environment variables.

This allows the CLI to automatically populate the appropriate flags:
- `RABBITMQ_ENDPOINT`: RabbitMQ server address to connect to.
- `RABBITMQ_HTTP_API_ENDPOINT`: RabbitMQ HTTP API server address (optional, please specify if different from the RabbitMQ server).

#### Example to set environment variables:

```bash
export RABBITMQ_ENDPOINT=amqp://username:password@localhost:5672/vhost
export RABBITMQ_HTTP_API_ENDPOINT=http://custom-api-server:15672
```

If the environment variables are not set, you'll need to explicitly specify the flags for each run:
```bash
./cli --endpoint amqp://username:password@localhost:5672/vhost --http-api-endpoint http://custom-api-server:15672 ...
```

## 🚀 How to Run

### Using pre-built binaries

1. Download the latest release from the **[releases page](https://github.com/happening-oss/rabbitmq-message-ops/releases)**.
2. Extract the archive.
3. Run: `./cli` to start the CLI.

### Building from source

1. Clone the repository: `git clone <repo-url>`
2. Install dependencies: `go mod download`
3. Run: `go build -o cli ./cmd/cli/` to build the CLI or `go run ./cmd/cli/` to start the CLI without building
4. Run: `./cli` to start the CLI.

In case of errors during queue management, please refer to the **[Error Recovery](#-error-recovery)** section.

## 🧰 Commands

Please use `./cli --help` to see the global flags and available commands with their respective flags.

### 👀 View

Retrieve messages from a queue:

```bash
./cli -q <srcQueueName> -f <filter-expression> view -c <count> -o <outputFile>
```

Example output:
```bash
{"headers":{"key":"value-0"},"contentType":"application/json","contentEncoding":"utf-8","deliveryMode":2,"correlationID":"correlation-0","replyTo":"response-queue","expiration":"60000","messageID":"msg-0","timestamp":"2024-12-18T17:03:21+01:00","type":"type-0","userID":"guest","appID":"app-0","routingKey":"admin.api","body":"{\"message\":\"This is message number 0\"}"}
{"headers":{"key":"value-1"},"contentType":"application/json","contentEncoding":"utf-8","deliveryMode":2,"priority":1,"correlationID":"correlation-1","replyTo":"response-queue","expiration":"60000","messageID":"msg-1","timestamp":"2024-12-18T17:03:21+01:00","type":"type-1","userID":"guest","appID":"app-1","routingKey":"admin.api","body":"{\"message\":\"This is message number 1\"}"}
{"headers":{"key":"value-2"},"contentType":"application/json","contentEncoding":"utf-8","deliveryMode":2,"priority":2,"correlationID":"correlation-2","replyTo":"response-queue","expiration":"60000","messageID":"msg-2","timestamp":"2024-12-18T17:03:21+01:00","type":"type-2","userID":"guest","appID":"app-2","routingKey":"admin.api","body":"{\"message\":\"This is message number 2\"}"}
```

View doesn't alter the queue so it can also be used as a **dry-run**. This is useful for testing the filter expression to ensure
it matches the intended messages before executing e.g. purge command.

Use `--count` flag as a number of messages to preview. Also, use `--verbosity info` to output the processed and selected messages count.

Example output:

```bash
./cli -q admin.api -f 'messageID matches "."' --verbosity info view --count=2
time=2024-12-23T18:00:22.219+01:00 level=INFO msg="source queue messages info" total=106 ready=106 unacknowledged=0
time=2024-12-23T18:00:22.220+01:00 level=INFO msg="processing source queue"
{"headers":{"key":"value-610"},"contentType":"application/json","contentEncoding":"utf-8","deliveryMode":2,"correlationID":"correlation-610","replyTo":"response-queue","messageID":"msg-610","timestamp":"2024-12-23T11:03:14+01:00","type":"type-610","userID":"guest","appID":"app-610","redelivered":true,"routingKey":"amq.gen-EW5Tg3EfvWnHY0XQqGYNrw","body":"{\"message\":\"This is message number 610\"}"}
{"headers":{"key":"value-611"},"contentType":"application/json","contentEncoding":"utf-8","deliveryMode":2,"priority":1,"correlationID":"correlation-611","replyTo":"response-queue","messageID":"msg-611","timestamp":"2024-12-23T11:03:14+01:00","type":"type-611","userID":"guest","appID":"app-611","routingKey":"amq.gen-EW5Tg3EfvWnHY0XQqGYNrw","body":"{\"message\":\"This is message number 611\"}"}
time=2024-12-23T18:00:23.924+01:00 level=INFO msg="moving messages from temporary to source queue"
time=2024-12-23T18:00:25.581+01:00 level=INFO msg="moving messages from temporary to source queue finished" movedMessages=106 duration=1.657099171s
time=2024-12-23T18:00:25.581+01:00 level=INFO msg="processing source queue finished" processedMessages=106 selectedMessages=106 duration=3.360623593s
```

### 🚚 Move

Move selected messages from one queue to another:

```bash
./cli -q <srcQueueName> -f <filter-expression> move -d <destQueueName>
```

### 📑 Copy

Copy selected messages from one queue to another:

```bash
./cli -q <srcQueueName> -f <filter-expression> copy -d <destQueueName>
```

### 🧹 Purge

Remove messages from a queue based on a filter:

```bash
./cli -q <srcQueueName> -f <filter-expression> purge
```

## 🔍 Filtering

Flexible message filtering based on message properties with filter expression (**[expr-lang](https://expr-lang.org/docs/language-definition)**).
The following fields are supported for filtering:
- **headers**: Message headers
- **contentType**: Content type of the message
- **contentEncoding**: Content encoding of the message
- **deliveryMode**: Delivery mode of the message
- **priority**: Priority of the message
- **correlationID**: Correlation ID of the message
- **replyTo**: Reply-to address
- **expiration**: Expiration time of the message
- **messageID**: Message ID
- **timestamp**: Timestamp of the message
- **type**: Type of the message
- **userID**: User ID associated with the message
- **appID**: Application ID associated with the message
- **redelivered**: Whether the message was redelivered
- **exchange**: Exchange associated with the message
- **routingKey**: Routing key used for the message

### Examples:

For testing purposes, you can run the **[rabbitmq.http](./rabbitmq.http)** (specify vhost and routing_key) to create a sample message. Successfully routed message should return `"routed": true`.

```bash
./cli -q <srcQueueName> -f 'messageID == "17342"' view
./cli -q <srcQueueName> -f 'type matches "^msg.*"' view
./cli -q <srcQueueName> -f 'timestamp > "2001-01-01T00:00:00Z" and timestamp < "2001-01-02T00:00:00Z"' view
./cli -q <srcQueueName> -f 'headers.key1 == "val1.123" and correlationID == "157"' view
./cli -q <srcQueueName> -f 'priority > 0' view
```

## 🚨 Error Recovery

In case of errors, please follow the instructions provided in the error message.
You will have all the necessary information to recover from the error.

### Queue

#### Partial queue management failure (e.g. move failed after the n-th message):
- Check if some messages have been moved from the source queue to temporary queue.
- Check if the last processed message (the one that caused the error, you can do that using "view --count=1") is requeued back to the front of the source queue.
  If message is not at the front, please move message to the front of the source queue manually.
  In case that acknowledgment failed but publishing to the destination queue (move/copy commands) succeeded, please manually remove the duplicated message from the source or destination queue.
- If some messages have been moved and last processed message is at the front, try to manage queue again and specify the --tempQueue parameter with the currently used temporary queue.
  That will cause QueueManager to continue processing from the last processed message that caused error and move all tempQueue messages (also those that were moved to tempQueue during the failed command) to source queue when finished, preserving the order.
#### Temporary to source queue failure (e.g. move from temporary to source queue failed):
- Check if the last processed message (the one that caused the error, you can do that using "view --count=1") is requeued back to the front of the temporary queue.
  If message is not at the front, please move message to the front of the temporary queue manually.
  In case that acknowledgment failed but publishing to the source queue succeeded, please manually remove the duplicated message from the temporary or source queue.
- Manually move remaining messages from the temporary queue to source queue (use "move" command).

### Stream

#### Partial stream processing failure (e.g. copy failed after the n-th message):
- Check if some messages have been copied to the destination queue.
- Act accordingly on the destination queue and try to manage stream again.

### SIGINT (Ctrl+C) and SIGTERM
You may also cancel the command execution by sending a **SIGINT** or **SIGTERM** signal. The CLI will gracefully stop the execution and provide instructions on how to recover.

## 📜 Ordering

> ⚠️ To preserve the order, each command processes **ALL** messages in the source queue which can be **time-consuming** for large queues.

To maintain the order of messages, the CLI tool moves messages to temporary queue and then back to the source queue.

### Example Scenario:

    Using the CLI tool:
        Command: view --count 3
        Effect: Order in the queue remains unchanged.

    Illustration (Msg1 is at the front of the queue):
        Initial Queue: [Msg5, Msg4, Msg3, Msg2, Msg1]
        After Viewing: [Msg5, Msg4, Msg3, Msg2, Msg1] (unchanged)

    Using RabbitMQ UI:
        Action: Viewing the first 3 messages with requeue.
        Effect: Messages are requeued at the end of the queue.

    Illustration (Msg1 is at the front of the queue):
        Initial Queue: [Msg5, Msg4, Msg3, Msg2, Msg1]
        After Viewing: [Msg3, Msg2, Msg1, Msg5, Msg4] (changed)

## 🤝 Contributing

Contributions are welcome. Please follow the existing code style and ensure all tests pass before submitting.

## 📞 Contact

If you encounter issues or have feature suggestions, please open an **[issue](https://github.com/happening-oss/rabbitmq-message-ops/issues)** on the GitHub repository.