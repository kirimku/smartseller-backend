# MockServer for Kirimku API Testing

This directory contains the configuration for a MockServer instance that simulates external API endpoints used by Kirimku Backend.

## Current Mock Implementations

- **JNT API (J&T Express)**: All shipping-related endpoints (tariff inquiry, booking, cancellation)
- Additional courier APIs can be added in the future (JNE, NinjaVan, SiCepat, etc.)

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- Bash shell (for the management script)

### Using the MockServer

1. Start the MockServer:
   ```
   ./mockserver_manager.sh start
   ```

2. Check if MockServer is running:
   ```
   ./mockserver_manager.sh status
   ```

3. Reset all expectations (useful when tests modify the mock behavior):
   ```
   ./mockserver_manager.sh reset
   ```

4. Stop the MockServer:
   ```
   ./mockserver_manager.sh stop
   ```

## Running Tests with MockServer

To run tests against the MockServer instead of real APIs, set the appropriate environment variables:

```bash
# Example: Running JNT tests against MockServer
JNT_API_URL=http://localhost:1080 go test -v ./internal/infrastructure/external/jnt_client_test.go
```

## Adding New Mock Endpoints

1. Edit the `mockserver-expectations.json` file to add new endpoint expectations
2. Restart the MockServer or run `./mockserver_manager.sh reset` to apply changes

## Viewing and Debugging MockServer

MockServer provides a dashboard UI at http://localhost:1080/mockserver/dashboard

You can use this UI to:
- View active expectations
- See recorded requests
- Verify mock responses
- Modify expectations on the fly

## Extending Mocks

To add mocks for additional APIs:

1. Create new expectation entries in `mockserver-expectations.json`
2. Follow the MockServer expectation format documented here: 
   https://www.mock-server.com/mock_server/creating_expectations.html

## Advanced Usage

For more detailed MockServer usage, see the official documentation:
https://www.mock-server.com/