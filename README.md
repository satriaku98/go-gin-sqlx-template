# Go Gin SQLX Template

A production-ready REST API template built with Go, Gin web framework, and SQLX for PostgreSQL database operations. This template follows clean architecture principles with clear separation of concerns.

## Features

- ✅ **Clean Architecture**: Separation of concerns across layers (delivery, usecase, repository)
- ✅ **Gin Web Framework**: Fast HTTP router and middleware support
- ✅ **SQLX**: Type-safe SQL operations with PostgreSQL
- ✅ **Viper Configuration**: Environment-based configuration management
- ✅ **Database Migrations**: Using golang-migrate
- ✅ **Manual Dependency Injection**: Full control over dependencies
- ✅ **Structured Logging**: Custom logger with different log levels
- ✅ **Standardized Responses**: Consistent API response format
- ✅ **Graceful Shutdown**: Proper cleanup on application termination
- ✅ **CORS Support**: Cross-origin resource sharing middleware
- ✅ **Request Logging**: HTTP request/response logging
- ✅ **Panic Recovery**: Automatic recovery from panics

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── config/
│   └── config.go                # Configuration management
├── internal/
│   ├── delivery/
│   │   └── http/
│   │       ├── handler/         # HTTP handlers
│   │       ├── middleware/      # HTTP middleware
│   │       └── router/          # Route definitions
│   ├── model/                   # Domain models and DTOs
│   ├── repository/              # Data access layer
│   │   └── postgres/            # PostgreSQL implementations
│   └── usecase/                 # Business logic layer
│       └── impl/                # Usecase implementations
├── pkg/
│   ├── database/                # Database connection
│   ├── logger/                  # Logging utility
│   └── utils/                   # Utility functions
├── migrations/                  # Database migrations
├── .env                         # Environment variables
├── go.mod                       # Go module definition
└── README.md                    # This file
```

## Prerequisites

- Go 1.25.2 or higher
- PostgreSQL 12 or higher
- golang-migrate CLI (for running migrations)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-gin-sqlx-template
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Create PostgreSQL database:
```bash
createdb go_gin_sqlx_db
```

5. Run database migrations:
```bash
# Install golang-migrate if not already installed
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/go_gin_sqlx_db?sslmode=disable" up
```

## Running the Application

```bash
go run cmd/api/main.go
```

The server will start on `http://localhost:8080` (or the port specified in your `.env` file).

## API Endpoints

### Health Check
```
GET /health
```

### User Management

#### Create User
```
POST /api/v1/users
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "John Doe",
  "password": "securepassword"
}
```

#### Get All Users (with pagination)
```
GET /api/v1/users?page=1&limit=10
```

#### Get User by ID
```
GET /api/v1/users/:id
```

#### Update User
```
PUT /api/v1/users/:id
Content-Type: application/json

{
  "email": "newemail@example.com",
  "name": "Jane Doe"
}
```

#### Delete User
```
DELETE /api/v1/users/:id
```

## Configuration

Configuration is managed through environment variables in the `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| `ENVIRONMENT` | Application environment | `development` |
| `SERVER_PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | PostgreSQL user | `postgres` |
| `DB_PASSWORD` | PostgreSQL password | `postgres` |
| `DB_NAME` | PostgreSQL database name | `go_gin_sqlx_db` |

## Database Migrations

### Create a new migration
```bash
migrate create -ext sql -dir migrations -seq <migration_name>
```

### Run migrations
```bash
migrate -path migrations -database "postgresql://user:password@localhost:5432/dbname?sslmode=disable" up
```

### Rollback migrations
```bash
migrate -path migrations -database "postgresql://user:password@localhost:5432/dbname?sslmode=disable" down
```

## Architecture

This template follows **Clean Architecture** principles:

1. **Delivery Layer** (`internal/delivery/http`): Handles HTTP requests, validation, and response formatting
2. **Usecase Layer** (`internal/usecase`): Contains business logic and orchestrates data flow
3. **Repository Layer** (`internal/repository`): Manages data access and database operations
4. **Model Layer** (`internal/model`): Defines domain entities and DTOs
5. **Infrastructure Layer** (`pkg/`): Provides cross-cutting concerns (database, logging, utilities)

### Dependency Flow
```
Handler → Usecase → Repository → Database
```

Dependencies are manually injected in `cmd/api/main.go` for better control and testability.

## Development

### Adding a New Entity

1. Create model in `internal/model/<entity>.go`
2. Define repository interface in `internal/repository/<entity>_repository.go`
3. Implement repository in `internal/repository/postgres/<entity>_repository.go`
4. Define usecase interface in `internal/usecase/<entity>_usecase.go`
5. Implement usecase in `internal/usecase/impl/<entity>_usecase_impl.go`
6. Create handler in `internal/delivery/http/handler/<entity>_handler.go`
7. Register routes in `internal/delivery/http/router/router.go`
8. Wire dependencies in `cmd/api/main.go`

## Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## Building for Production

```bash
# Build binary
go build -o bin/api cmd/api/main.go

# Run binary
./bin/api
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.