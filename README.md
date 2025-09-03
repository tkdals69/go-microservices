# Go Microservices Project

This project is a collection of microservices built using Go. It follows a cloud-neutral architecture and adheres to the 12-factor app principles. The microservices are designed to be modular, scalable, and maintainable.

## Project Structure

```
go-microservices
├── cmd
│   ├── gateway         # Gateway microservice
│   ├── progression     # Progression microservice
│   ├── leaderboard     # Leaderboard microservice
│   └── fairness        # Fairness microservice
├── pkg
│   ├── config          # Configuration management
│   ├── adapters        # Cloud, database, and cache adapters
│   ├── middleware      # Middleware for security and observability
│   ├── observability    # Health checks and metrics
│   ├── handlers        # HTTP handlers for each microservice
│   └── tests          # Unit and integration tests
├── api                 # OpenAPI specifications
├── .env                # Environment variables
├── Makefile            # Build, run, and test commands
├── README.md           # Project documentation
├── LICENSE             # Licensing information
├── go.mod              # Module dependencies
├── go.sum              # Dependency checksums
├── Dockerfile.gateway   # Docker configuration for gateway
├── Dockerfile.progression # Docker configuration for progression
├── Dockerfile.leaderboard # Docker configuration for leaderboard
└── Dockerfile.fairness  # Docker configuration for fairness
```

## Getting Started

### Prerequisites

- Go 1.22 or later
- Docker (for containerization)
- PostgreSQL (for database)
- Redis (for caching)

### Installation

1. Clone the repository:
   ```
   git clone <repository-url>
   cd go-microservices
   ```

2. Set up environment variables:
   Create a `.env` file in the root directory and configure your environment variables.

3. Build the project:
   ```
   make build
   ```

### Running the Services

You can run each microservice individually using the following commands:

- For Gateway:
  ```
  make run-gateway
  ```

- For Progression:
  ```
  make run-progression
  ```

- For Leaderboard:
  ```
  make run-leaderboard
  ```

- For Fairness:
  ```
  make run-fairness
  ```

### Testing

To run the tests, use:
```
make test
```

### API Documentation

The API specifications are available in the `api/openapi.yaml` file. You can use tools like Swagger UI to visualize and interact with the API.

## Observability

Each microservice exposes the following endpoints for observability:

- `/healthz` - Health check endpoint
- `/metrics` - Metrics endpoint for Prometheus

## License

This project is licensed under the MIT License. See the LICENSE file for more details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements or features.