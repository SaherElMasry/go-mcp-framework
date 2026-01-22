# 🌤️ Weather MCP Server Example

**Production-ready weather server using MCP Framework v0.3.0**

This example demonstrates:
- ✅ Framework v0.3.0 auth integration
- ✅ Streaming tools (Server-Sent Events)
- ✅ Caching & rate limiting
- ✅ Prometheus metrics
- ✅ Health checks
- ✅ JSON structured logging

---

## 📁 Project Structure

```
examples/weather-server/
├── cmd/server/
│   └── main.go              # Server entry point
├── internal/weather/
│   └── backend.go           # Weather backend implementation
├── config/
│   └── config.yaml          # Server configuration
├── README.md                # This file
├── go.mod                   # Go module file
└── Makefile                 # Build & run commands
```

---

## 🚀 Quick Start

### 1. Get API Key

Sign up for free at [WeatherAPI.com](https://www.weatherapi.com/signup.aspx)

### 2. Set Environment Variable

```bash
export WEATHER_API_KEY="your-api-key-here"
```

### 3. Run Server

```bash
cd examples/weather-server
go run cmd/server/main.go
```

---

## 📖 Features

### 🔧 Available Tools

#### 1. **get_current_weather** (Non-streaming)
Get current weather conditions for a location.

**Parameters:**
- `location` (string, required): City name, zip code, or coordinates

**Example:**
```bash
curl -X POST http://localhost:8080 \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc":"2.0",
    "id":1,
    "method":"tools/call",
    "params":{
      "name":"get_current_weather",
      "arguments":{"location":"London"}
    }
  }'
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{
      "type": "text",
      "text": {
        "summary": "Current weather in London, England (United Kingdom):\n🌡️ Temperature: 15.0°C (Feels like 14.0°C)\n☁️ Condition: Partly cloudy\n💧 Humidity: 72%\n🌬️ Wind: 15.0 kph coming from WSW\n🕒 Last Updated: 2024-01-20 10:30",
        "raw_data": {...}
      }
    }]
  }
}
```

---

#### 2. **get_forecast** (Non-streaming)
Get multi-day weather forecast (1-10 days).

**Parameters:**
- `location` (string, required): City name, zip code, or coordinates
- `days` (integer, optional): Number of days (1-10, default: 3)

**Example:**
```bash
curl -X POST http://localhost:8080 \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc":"2.0",
    "id":1,
    "method":"tools/call",
    "params":{
      "name":"get_forecast",
      "arguments":{"location":"Paris","days":5}
    }
  }'
```

---

#### 3. **search_locations** (Streaming 🌊)
Search for locations with real-time progress.

**Parameters:**
- `query` (string, required): Search query

**Example (SSE):**
```bash
curl -N http://localhost:8080/stream \
  -H 'Content-Type: application/json' \
  -d '{
    "tool":"search_locations",
    "arguments":{"query":"London"}
  }'
```

**Stream Output:**
```
event: progress
data: {"current":30,"total":100,"message":"Querying WeatherAPI..."}

event: data
data: {"index":1,"total":3,"location":{...}}

event: progress
data: {"current":100,"total":100,"message":"Processed 3/3 locations"}

event: done
data: {"status":"completed"}
```

---

#### 4. **get_astronomy** (Non-streaming)
Get sunrise, sunset, and moon phase data.

**Parameters:**
- `location` (string, required): City name
- `date` (string, optional): Date in YYYY-MM-DD format (default: today)

**Example:**
```bash
curl -X POST http://localhost:8080 \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc":"2.0",
    "id":1,
    "method":"tools/call",
    "params":{
      "name":"get_astronomy",
      "arguments":{"location":"Tokyo","date":"2024-01-20"}
    }
  }'
```

**Response:**
```
🔭 Astronomy Report for Tokyo, Japan (2024-01-20):
☀️ Sun: Sunrise at 06:51 AM, Sunset at 04:55 PM
🌙 Moon: Waxing Gibbous Phase with 75% illumination. Rises at 01:15 PM, Sets at 04:30 AM
```

---

#### 5. **bulk_weather_check** (Streaming 🌊)
Get weather for multiple locations with progress tracking.

**Parameters:**
- `locations` (string, required): Comma-separated list of cities

**Example:**
```bash
curl -N http://localhost:8080/stream \
  -H 'Content-Type: application/json' \
  -d '{
    "tool":"bulk_weather_check",
    "arguments":{"locations":"London,Paris,Tokyo,New York"}
  }'
```

---

## 🔐 Authentication (Framework v0.3.0)

This example uses the **framework's built-in auth system**:

### In Backend (internal/weather/backend.go):
```go
// Setup API key provider using framework auth
apiKeyProvider := auth.NewAPIKeyProvider("weather-api", auth.APIKeyConfig{
    APIKey: b.apiKey,
    Header: "key", // WeatherAPI uses 'key' parameter
})

// Register resource
apiKeyProvider.RegisterResource(auth.ResourceConfig{
    ID:   "weather-api",
    Type: "http",
    Config: map[string]interface{}{
        "base_url": b.baseURL,
    },
})

// Set on backend
b.SetAuthProvider(apiKeyProvider)
```

### Benefits:
- ✅ Centralized auth management
- ✅ Automatic credential injection
- ✅ Thread-safe operations
- ✅ Metrics tracking
- ✅ Error handling

---

## 📊 Monitoring & Observability

### Prometheus Metrics
```bash
curl http://localhost:9091/metrics
```

**Available metrics:**
- `mcp_tool_calls_total` - Total tool calls
- `mcp_tool_call_duration_seconds` - Tool execution time
- `mcp_streaming_events_total` - Streaming events emitted
- `mcp_errors_total` - Error count

### Health Check
```bash
curl http://localhost:9091/health
```

### Logs
Structured JSON logging to stdout:
```json
{
  "time":"2024-01-20T10:30:00Z",
  "level":"INFO",
  "msg":"tool called",
  "tool":"get_current_weather",
  "location":"London",
  "duration_ms":245
}
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Weather MCP Server                      │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │   HTTP       │  │  Streaming  │  │   Observability  │  │
│  │  Transport   │  │   Engine    │  │    (Metrics)     │  │
│  └──────┬───────┘  └──────┬──────┘  └────────┬─────────┘  │
│         │                  │                  │            │
│         └──────────┬───────┴──────────────────┘            │
│                    │                                        │
│         ┌──────────▼───────────┐                           │
│         │   Protocol Handler   │                           │
│         └──────────┬───────────┘                           │
│                    │                                        │
│         ┌──────────▼───────────┐                           │
│         │   Weather Backend    │                           │
│         └──────────┬───────────┘                           │
│                    │                                        │
│         ┌──────────▼───────────┐                           │
│         │   Auth Provider      │ (Framework v0.3.0)        │
│         │   (API Key)          │                           │
│         └──────────┬───────────┘                           │
│                    │                                        │
│                    ▼                                        │
│         ┌─────────────────────┐                            │
│         │   WeatherAPI.com    │                            │
│         └─────────────────────┘                            │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧪 Testing

### Test List Tools
```bash
curl -X POST http://localhost:8080 \
  -H 'Content-Type: application/json' \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

### Test Current Weather
```bash
curl -X POST http://localhost:8080 \
  -H 'Content-Type: application/json' \
  -d '{
    "jsonrpc":"2.0",
    "id":1,
    "method":"tools/call",
    "params":{
      "name":"get_current_weather",
      "arguments":{"location":"London"}
    }
  }'
```

### Test Streaming
```bash
curl -N http://localhost:8080/stream \
  -H 'Content-Type: application/json' \
  -d '{
    "tool":"search_locations",
    "arguments":{"query":"New York"}
  }'
```

---

## 🔧 Configuration

Edit `config/config.yaml` to customize:

```yaml
backend:
  config:
    api_key: "${WEATHER_API_KEY}"
    timeout: 30s
    cache_ttl: 300

streaming:
  enabled: true
  buffer_size: 100
  max_concurrent: 8

observability:
  enabled: true
  metrics_address: ":9091"

logging:
  level: "info"
  format: "json"
```

---

## 📝 Key Differences from Local Auth

| Aspect | Local pkg/auth | Framework v0.3.0 Auth |
|--------|----------------|----------------------|
| **Location** | `pkg/auth` in your project | Framework's `auth` package |
| **Setup** | Manual provider creation | Integrated with server |
| **Resources** | Manual registration | `WithAuthResource()` option |
| **Management** | Manual lifecycle | Automatic lifecycle |
| **Metrics** | Custom implementation | Built-in metrics |
| **Testing** | Custom mocks | Framework mocks |

---

## 🚀 Deployment

### Docker
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go
ENV WEATHER_API_KEY=""
EXPOSE 8080 9091
CMD ["./server"]
```

### Environment Variables
```bash
WEATHER_API_KEY=your-key-here
HTTP_ADDRESS=:8080
METRICS_ADDRESS=:9091
LOG_LEVEL=info
```

---

## 📚 Learn More

- [MCP Framework Documentation](../../README.md)
- [Auth System Guide](../../auth/README.md)
- [Streaming Guide](../../docs/STREAMING.md)
- [WeatherAPI Documentation](https://www.weatherapi.com/docs/)

---

## 🎯 Next Steps

1. ✅ Try the example
2. ✅ Understand the auth integration
3. ✅ Add your own tools
4. ✅ Deploy to production!

---

**Built with ❤️ using MCP Framework v0.3.0**
