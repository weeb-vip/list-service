# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build & Run
- `go run cmd/main.go serve` - Start the GraphQL API server
- `go build -o main cmd/main.go` - Build the binary
- `docker-compose up` - Start dependencies (MySQL, Redis)

### Testing
- `go test ./...` - Run all tests
- `go test -v ./http/handlers/...` - Run specific handler tests

### Code Generation
- `make gql` or `go run github.com/99designs/gqlgen generate` - Generate GraphQL code
- `make generate` - Generate mocks and GraphQL code
- `make mocks` - Generate mock files

### Database Operations
- `go run cmd/main.go migrate up` - Run database migrations
- `make migrate` - Run migrations (alias)
- `make create-migration name=migration_name` - Create new migration
- `migrate create -ext sql -dir db/migrations -seq migration_name` - Alternative migration creation

## Architecture Overview

This is a Go-based GraphQL microservice for anime list management with the following structure:

### Core Components
- **GraphQL API**: Built with gqlgen, provides federated GraphQL schema
- **Database**: MySQL with GORM as ORM, managed migrations in `db/migrations/`
- **Message Queue**: Apache Pulsar integration for event publishing
- **Metrics**: Prometheus and DataDog integration via `weeb-vip/go-metrics-lib`

### Directory Structure
- `cmd/` - Application entry point with Cobra CLI commands
- `graph/` - GraphQL schema definitions, resolvers, and generated code
- `internal/` - Core business logic:
  - `commands/` - CLI command implementations (serve, migrate, etc.)
  - `db/` - Database connection and repository pattern implementations
  - `resolvers/` - GraphQL resolver implementations
  - `services/` - Business logic layer
- `http/` - HTTP server setup, middleware (logging, CORS, request info)
- `metrics/` - Metrics collection and publishing

### Key Features
- **User Lists**: CRUD operations for anime lists per user
- **User Anime**: Track anime in lists with status, progress, ratings
- **Authentication**: Uses `@Authenticated` directive (external auth service)
- **Federation**: GraphQL federation support with `@key` directives

### Database Entities
- `user_list` - User's anime lists
- `user_anime` - Anime entries within lists with metadata

### Configuration
- Environment-specific configs in `config/` directory
- Docker containerization with distroless final image
- Runs on port 3000 in container

### Development Dependencies
- Go 1.23
- MySQL 8.0.33 (via docker-compose)
- Redis 5.0.5 (via docker-compose)