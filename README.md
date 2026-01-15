# BookCabin - Flight Search and Aggregation API

[![Go Version](https://img.shields.io/badge/Go-1.25.3-blue.svg)](https://golang.org)

BookCabin is a high-performance flight search and aggregation service built in Go. It provides a unified API to search flights across multiple airline providers, supporting both one-way, round-trip, and multi-city searches with intelligent result aggregation and scoring.

## 📑 Table of Contents

- [🚀 Features](#-features)
- [🏗️ Architecture](#️-architecture)
- [🚦 Quick Start](#-quick-start)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [🔧 Configuration](#-configuration)
- [📚 API Usage](#-api-usage)
  - [Flight Search](#flight-search)
  - [Response Format](#response-format)
  - [Health Check](#health-check)
- [🧪 Development](#-development)
  - [Available Mock Flight Data](#available-mock-flight-data)
  - [Running Tests](#running-tests)
  - [API Documentation](#api-documentation)
  - [Project Structure](#project-structure)
- [🏢 Production Deployment](#-production-deployment)
- [🐛 Troubleshooting](#-troubleshooting)
- [📊 Performance](#-performance)
- [👤 Author](#-author)
- [🙏 Acknowledgments](#-acknowledgments)

## 🚀 Features

- **Multi-Provider Aggregation**: Simultaneously searches across multiple airline providers (AirAsia, Batik Air, Garuda Indonesia, Lion Air)
- **Flexible Search Types**: 
  - One-way flights
  - Round-trip flights  
  - Multi-city itineraries
- **Intelligent Scoring**: Built-in flight scoring algorithm based on price, duration, and stops
- **Resilient Architecture**: Retry logic, timeout handling, and graceful failure management
- **RESTful API**: Clean HTTP API with Swagger documentation
- **Comprehensive Testing**: Extensive unit test coverage with mock providers
- **Production Ready**: Structured logging, metrics, CORS support, and configurable timeouts

## 🏗️ Architecture

```
┌─────────────────┐
│   HTTP Client   │
└─────────┬───────┘
          │
┌─────────▼───────┐
│  HTTP Router    │  ← Chi Router + Middleware
└─────────┬───────┘
          │
┌─────────▼───────┐
│   Aggregator    │  ← Flight Search Orchestrator
└─────────┬───────┘
          │
┌─────────▼───────┐
│   Providers     │  ← AirAsia, Batik, Garuda, Lion
└─────────────────┘
```

### Core Components

- **HTTP Layer** (`api/http/`): REST API endpoints and middleware
- **Aggregator Service** (`service/aggregator/`): Orchestrates searches across providers
- **Provider Services** (`service/{airasia,batik,garuda,lion}/`): Individual airline integrations
- **Configuration** (`configs/`): Environment-based configuration management
- **Utilities** (`pkg/`): Shared packages for logging, validation, response handling

## 🚦 Quick Start

### Prerequisites

- Go 1.25.3 or later
- Make (optional, for convenience commands)
- Swag CLI for API documentation generation: `go install github.com/swaggo/swag/cmd/swag@latest`

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/elkoshar/bookcabin.git
   cd bookcabin
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up configuration**
   ```bash
   cp configs/.env.sample configs/.env
   # Edit configs/.env with your settings
   ```

4. **Build and run**
   ```bash
   make run-http
   ```

   Or manually:
   ```bash
   go build -o bin/bookcabin-http cmd/http/main.go
   ./bin/bookcabin-http
   ```

   Or run directly from IDE/terminal:
   ```bash
   # Generate Swagger docs first (optional, only if there are API changes)
   swag init --parseDependency --parseInternal --parseDepth 2 -g cmd/http/main.go
   
   # Run directly without building
   go run cmd/http/main.go
   ```

The service will start on `http://localhost:8080` by default.

**Access Swagger Documentation:** Once running, visit `http://localhost:8080/swagger/index.html` to view the interactive API documentation.

## 🔧 Configuration

Configuration is managed through environment variables. Copy `configs/.env.sample` to `configs/.env` and customize:

```env
# Server Configuration
SERVER_PORT=8080
SERVER_SHUTDOWN_TIMEOUT=10
LOG_LEVEL=INFO
ENV=development

# Timeouts
HTTP_INBOUND_TIMEOUT=60s
AGGREGATOR_TIMEOUT=10s

# Provider Mock Data Paths
GARUDA_PATH=mock_data/garuda_indonesia_search_response.json
LION_PATH=mock_data/lion_air_search_response.json
AIRASIA_PATH=mock_data/airasia_search_response.json
BATIK_PATH=mock_data/batik_air_search_response.json
```

## 📚 API Usage

### Flight Search

**Endpoint:** `POST /bookcabin/flight/search`

**Headers:**
```
Content-Type: application/json
Accept-Language: en
```

#### One-way Flight
```json
{
  "origin": "CGK",
  "destination": "DPS", 
  "departure_date": "2025-12-15",
  "passengers": 1,
  "cabin_class": "economy"
}
```

#### Round-trip Flight
```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departure_date": "2025-12-15", 
  "return_date": "2025-12-17",
  "passengers": 2,
  "cabin_class": "economy"
}
```

#### Multi-city Flight
```json
{
  "passengers": 1,
  "cabin_class": "economy",
  "segments": [
    {
      "origin": "CGK",
      "destination": "DPS", 
      "departure_date": "2025-12-15"
    },
    {
      "origin": "DPS",
      "destination": "SUB",
      "departure_date": "2025-12-17"
    }
  ]
}
```

### Response Format

```json
{
  "code": 200,
  "data": {
    "search_criteria": {
      "origin": "CGK",
      "destination": "DPS",
      "departure_date": "2025-12-15",
      "passengers": 1,
      "cabin_class": "economy"
    },
    "metadata": {
      "total_results": 12,
      "providers_queried": 4,
      "providers_succeeded": 4,
      "providers_failed": 0,
      "search_time_ms": 856
    },
    "flights": [
      {
        "id": "QZ520_CGK_DPS_20251215",
        "provider": "AirAsia",
        "airline": {
          "name": "AirAsia",
          "code": "QZ"
        },
        "flight_number": "QZ520", 
        "departure": {
          "airport": "CGK",
          "city": "Jakarta",
          "datetime": "2025-12-15T04:45:00+07:00",
          "timestamp": 1734234300
        },
        "arrival": {
          "airport": "DPS", 
          "city": "Denpasar",
          "datetime": "2025-12-15T07:25:00+08:00",
          "timestamp": 1734243900
        },
        "duration": {
          "total_minutes": 100,
          "formatted": "1h 40m"
        },
        "stops": 0,
        "price": {
          "amount": 650000,
          "currency": "IDR",
          "formatted": "Rp 650.000"
        },
        "available_seats": 67,
        "cabin_class": "economy",
        "amenities": ["wifi", "meal"],
        "score": 2.34
      }
    ],
    "return_flights": [], // For round-trip searches
    "multi_city_flights": [] // For multi-city searches  
  }
}
```

### Health Check

**Endpoint:** `GET /bookcabin/health`

```json
{
  "code": 200,
  "message": "OK"
}
```

## 🧪 Development

### Available Mock Flight Data

The service includes mock data for testing. Here are the available flights for each provider:

#### AirAsia Flights:
| Flight Code | Origin | Destination | Departure Date |
|-------------|--------|-------------|----------------|
| QZ520 | CGK | DPS | 2025-12-15T04:45:00+07:00 |
| QZ524 | CGK | DPS | 2025-12-15T10:00:00+07:00 |
| QZ532 | CGK | DPS | 2025-12-15T19:30:00+07:00 |
| QZ7250 | CGK | DPS | 2025-12-15T15:15:00+07:00 |
| QZ7760 | CGK | SUB | 2025-12-15T08:00:00+07:00 |
| QZ7510 | CGK | DPS | 2025-12-16T10:00:00+07:00 |
| QZ7520 | CGK | DPS | 2025-12-20T14:00:00+07:00 |
| QZ7771 | SUB | CGK | 2025-12-15T18:00:00+07:00 |
| QZ7541 | DPS | CGK | 2025-12-17T09:00:00+08:00 |

#### Batik Air Flights:
| Flight Code | Origin | Destination | Departure Date |
|-------------|--------|-------------|----------------|
| ID6514 | CGK | DPS | 2025-12-15T07:15:00+0700 |
| ID6520 | CGK | DPS | 2025-12-15T13:30:00+0700 |
| ID7042 | CGK | DPS | 2025-12-15T18:45:00+0700 |
| ID6870 | CGK | SUB | 2025-12-15T09:00:00+0700 |
| ID6508 | CGK | DPS | 2025-12-16T11:00:00+0700 |
| ID6518 | CGK | DPS | 2025-12-20T13:00:00+0700 |
| ID6873 | SUB | CGK | 2025-12-15T21:00:00+0700 |
| ID6529 | DPS | CGK | 2025-12-17T12:00:00+0800 |

#### Garuda Indonesia Flights:
| Flight Code | Origin | Destination | Departure Date |
|-------------|--------|-------------|----------------|
| GA400 | CGK | DPS | 2025-12-15T06:00:00+07:00 |
| GA410 | CGK | DPS | 2025-12-15T09:30:00+07:00 |
| GA315 | CGK | SUB | 2025-12-15T14:00:00+07:00 |
| GA312 | CGK | SUB | 2025-12-15T07:00:00+07:00 |
| GA402 | CGK | DPS | 2025-12-16T09:00:00+07:00 |
| GA412 | CGK | DPS | 2025-12-20T12:00:00+07:00 |
| GA320 | SUB | CGK | 2025-12-15T19:00:00+07:00 |
| GA415 | DPS | CGK | 2025-12-17T11:00:00+08:00 |

#### Lion Air Flights:
| Flight Code | Origin | Destination | Departure Date |
|-------------|--------|-------------|----------------|
| JT740 | CGK | DPS | 2025-12-15T05:30:00 |
| JT742 | CGK | DPS | 2025-12-15T11:45:00 |
| JT650 | CGK | DPS | 2025-12-15T16:20:00 |
| JT690 | CGK | SUB | 2025-12-15T06:00:00 |
| JT750 | CGK | DPS | 2025-12-16T08:00:00 |
| JT756 | CGK | DPS | 2025-12-20T11:00:00 |
| JT699 | SUB | CGK | 2025-12-15T20:00:00 |
| JT761 | DPS | CGK | 2025-12-17T10:00:00 |

**Airport Codes:**
- **CGK**: Soekarno-Hatta International Airport, Jakarta
- **DPS**: Ngurah Rai International Airport, Denpasar, Bali  
- **SUB**: Juanda International Airport, Surabaya

**Test Examples:**
- **Popular route**: CGK → DPS (Jakarta to Bali) on 2025-12-15
- **Return trip**: DPS → CGK on 2025-12-17
- **Alternative destination**: CGK → SUB (Jakarta to Surabaya)
- **Multi-city**: CGK → DPS → SUB or CGK → SUB → DPS

### Running Tests

```bash
# Run all tests
make test
# or
go test ./...

# Run tests with verbose output
go test ./... -v

# Run tests for specific package
go test ./service/aggregator/... -v

# Run specific test
go test ./service/aggregator/... -run TestFlightAggregator_SearchMultiCity -v
```

### API Documentation

Generate and view Swagger documentation:

```bash
# Generate docs
make swag

# Start server and visit
http://localhost:8080/swagger/index.html
```

### Project Structure

```
bookcabin/
├── api/                    # API layer
│   ├── http/
│   │   └── aggregator/     # HTTP handlers
│   ├── interface.go        # Service interfaces
│   └── middleware.go       # HTTP middleware
├── cmd/
│   └── http/              # Application entrypoints
├── configs/               # Configuration management
├── docs/                  # Swagger documentation
├── mock_data/            # Test data for providers
├── pkg/                  # Shared utilities
│   ├── helpers/          # Helper functions
│   ├── logger/           # Structured logging
│   ├── response/         # HTTP response handling
│   └── validator/        # Request validation
├── server/               # Server setup and DI
├── service/              # Business logic
│   ├── aggregator/       # Flight aggregation service
│   ├── airasia/         # AirAsia provider
│   ├── batik/           # Batik Air provider  
│   ├── garuda/          # Garuda Indonesia provider
│   └── lion/            # Lion Air provider
└── vendor/              # Dependencies
```

## 🏢 Production Deployment

### Docker

```bash
# Build image
make build-image-http

# Run container
make docker-run-http GO_ENV=production
```

### Environment Variables

For production deployment, ensure these environment variables are set:

```env
ENV=production
LOG_LEVEL=INFO
SERVER_PORT=8080
AGGREGATOR_TIMEOUT=10s
HTTP_INBOUND_TIMEOUT=60s
```

### Code Standards

- Follow Go best practices and conventions
- Write comprehensive unit tests (aim for >90% coverage)
- Use external test packages (`package service_test`)
- Document public APIs with Go comments
- Update Swagger documentation for API changes

## 🐛 Troubleshooting

### Common Issues

**Service won't start**
- Check if port 8080 is already in use
- Verify configuration in `configs/.env`
- Ensure mock data files exist

**Tests failing**
- Run `go mod download` to ensure dependencies
- Check that mock data files are present
- Verify Go version compatibility (1.25.3+)

**Slow response times**
- Adjust `AGGREGATOR_TIMEOUT` in configuration
- Check provider response times
- Consider reducing the number of concurrent providers

## 📊 Performance

The service is designed for high performance:

- **Concurrent provider searches**: All providers are queried simultaneously
- **Configurable timeouts**: Prevent slow providers from degrading overall performance  
- **Retry logic**: Built-in exponential backoff for transient failures
- **Intelligent scoring**: Results are sorted by a composite score algorithm
- **Memory efficient**: Minimal allocations in hot paths

## 👤 Author

**Elko Sharhadi Eppasa**
- GitHub: [@elkoshar](https://github.com/elkoshar)
- Email: elko.s.eppasa@mail.com
