# 🚀 Go S3 Sharing

[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/vchitai/go-s3-sharing)](https://goreportcard.com/report/github.com/vchitai/go-s3-sharing)
[![Build Status](https://github.com/vchitai/go-s3-sharing/workflows/CI/badge.svg)](https://github.com/vchitai/go-s3-sharing/actions)

A high-performance, secure, and scalable file sharing service built with Go that provides time-limited, authenticated access to S3 objects through Redis-based authorization.

## ✨ Features

- 🔐 **Secure Access**: Redis-based authentication with time-limited secrets
- ⏰ **Auto-Expiration**: Configurable link expiration (default: 90 days)
- 🚀 **High Performance**: Built with Go for maximum throughput and low latency
- 🏗️ **Clean Architecture**: Well-structured, testable, and maintainable codebase
- 🐳 **Docker Ready**: Complete containerization with Docker and Kubernetes support
- 📊 **Observability**: Built-in health checks, metrics, and structured logging
- 🛠️ **CLI Tool**: Easy-to-use command-line interface for generating shareable links
- 🔧 **Production Ready**: Comprehensive error handling, configuration management, and monitoring

## 🏃‍♂️ Quick Start

### Prerequisites

- Go 1.23+
- Redis server
- AWS S3 bucket
- AWS credentials configured

### Installation

```bash
# Clone the repository
git clone https://github.com/vchitai/go-s3-sharing.git
cd go-s3-sharing

# Install dependencies
go mod tidy

# Build the server
go build -o bin/server ./cmd/server

# Build the CLI tool
go build -o bin/cli ./cmd/cli
```

### Configuration

Set the following environment variables:

```bash
export S3_BUCKET="your-s3-bucket-name"
export AWS_REGION="us-east-1"
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""
export REDIS_DB="0"
export PORT="8080"
export MAX_AGE_DAYS="90"
```

### Running the Server

```bash
./bin/server
```

The server will start on port 8080 (or the port specified in the `PORT` environment variable).

### Using the CLI

Create a shareable link for an S3 object:

```bash
./bin/cli images/photo.jpg 24
```

This creates a link that expires in 24 hours.

## 🏗️ Architecture

The project follows Clean Architecture principles with clear separation of concerns:

```
├── cmd/                    # Application entry points
│   ├── server/            # HTTP server
│   └── cli/               # Command-line tool
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── domain/           # Business logic and entities
│   ├── service/          # Application services
│   └── transport/        # HTTP handlers and middleware
├── pkg/                   # Public library code (if any)
├── deployments/           # Docker and Kubernetes configs
├── examples/              # Usage examples
└── tests/                 # Integration and e2e tests
```

### Key Components

- **Domain Layer**: Core business logic and entities
- **Service Layer**: Application services implementing business use cases
- **Transport Layer**: HTTP handlers and API endpoints
- **Infrastructure Layer**: External dependencies (S3, Redis)

## 🔧 API Reference

### Endpoints

#### `GET /{yy}/{mm}/{dd}/{secret}/{path}`

Retrieves a shared file from S3.

**Path Parameters:**
- `yy/mm/dd`: Date when the link was created (YY-MM-DD format)
- `secret`: Authentication secret
- `path`: S3 object path

**Response:**
- `200 OK`: File content with appropriate Content-Type
- `400 Bad Request`: Invalid path or date format
- `401 Unauthorized`: Invalid or missing secret
- `403 Forbidden`: Link has expired
- `404 Not Found`: S3 object not found

**Note:** This endpoint uses a catch-all pattern and should be registered last in the router to avoid conflicts with other endpoints.

#### `POST /api/shares`

Creates a new shareable link.

**Request Body:**
```json
{
  "s3_path": "images/photo.jpg",
  "secret": "your-secret-key",
  "expires_at": "2024-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "url": "https://your-domain.com/24/12/31/secret/images/photo.jpg",
  "expires_at": "2024-12-31T23:59:59Z",
  "max_age_seconds": 86400
}
```

#### `GET /health`

Health check endpoint.

**Response:**
```json
{
  "status": "healthy"
}
```

#### `GET /ready`

Readiness check endpoint.

**Response:**
```json
{
  "status": "ready"
}
```

## 🐳 Docker Deployment

### Using Docker Compose

```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - S3_BUCKET=your-bucket
      - REDIS_ADDR=redis:6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

### Using Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

## 🧪 Testing

Run the test suite:

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Test routing fix specifically
go run ./examples/routing-test
```

## 📊 Monitoring

The service includes built-in monitoring capabilities:

- **Health Checks**: `/health` and `/ready` endpoints
- **Structured Logging**: JSON-formatted logs with context
- **Metrics**: Prometheus-compatible metrics (coming soon)
- **Tracing**: OpenTelemetry support (coming soon)

## 🔒 Security

- **Input Validation**: All inputs are validated and sanitized
- **Path Traversal Protection**: S3 paths are cleaned and validated
- **Time-based Expiration**: Links automatically expire
- **Secret-based Authentication**: Cryptographically secure secrets
- **Rate Limiting**: Built-in rate limiting (coming soon)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- [Go Redis](https://github.com/redis/go-redis)
- [Go Standard Library](https://golang.org/pkg/)

## 📞 Support

- 📧 Email: chitai.vct@gmail.com
- 🐛 Issues: [GitHub Issues](https://github.com/vchitai/go-s3-sharing/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/vchitai/go-s3-sharing/discussions)

---

Made with ❤️ by [Tai Vong](https://github.com/vchitai)