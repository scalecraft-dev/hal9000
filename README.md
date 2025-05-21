# HAL - Slack Incident Management Bot

HAL is a Slack bot designed to help teams manage incidents efficiently. It provides commands for creating, updating, and tracking incidents directly within Slack.

## Features

- Create incident channels with appropriate naming conventions
- Track incident timelines
- Manage action items
- Update incident status and severity
- Automatic postmortem creation for high-severity incidents

## Project Structure

The project follows a simple, flat structure with a single internal package:

```tree
/
├── main.go                 # Entry point and server setup
├── internal/               # Private application code
│   ├── models.go           # Domain models
│   ├── config.go           # Configuration management
│   ├── slack.go            # Slack API integration
│   ├── incident.go         # Incident management
│   ├── handlers.go         # HTTP handlers
│   └── routes.go           # Route registration
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- A Slack workspace with permissions to create a bot

### Environment Variables

Create a `.env` file with the following variables:

```bash
SLACK_TOKEN=xoxb-your-slack-token
SLACK_SIGNING_SECRET=your-slack-signing-secret
SERVER_PORT=50051
ENVIRONMENT=development
LOG_LEVEL=info
```

### Running the Application

```bash
# Build the application
make build

# Run the application
make run
```

### Slack Commands

- `/incident create` - Create a new incident
- `/incident update` - Update an existing incident
- `/incident timeline <message>` - Add an entry to the incident timeline
- `/incident action-item <description>` - Add an action item
- `/incident help` - Show available commands

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Linting

```bash
make lint
```

## License

[MIT](LICENSE)
