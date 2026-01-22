# 🔄 Migration Guide: Local Auth → Framework v0.3.0 Auth

This guide shows how to migrate from your local `pkg/auth` to the framework's built-in auth system.

---

## 📋 Summary of Changes

### What Changed?
1. ✅ Removed local `pkg/auth` package
2. ✅ Now using framework's `auth` package
3. ✅ Backend uses `SetAuthProvider()` instead of manual setup
4. ✅ Resources registered via framework options
5. ✅ Automatic lifecycle management

---

## 🔍 Side-by-Side Comparison

### **Before: Local pkg/auth**

```go
// internal/weather/backend.go
import (
    "weather-mcp-server/pkg/auth" // ← Local package
)

type WeatherBackend struct {
    *backend.BaseBackend
    authProvider auth.Provider // ← Local auth.Provider
    apiKey       string
}

func (b *WeatherBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
    // Manual auth setup
    b.authProvider = auth.NewAPIKeyProvider("weather-api", auth.APIKeyConfig{
        APIKey: b.apiKey,
    })
    
    // Manual resource registration
    if apiProvider, ok := b.authProvider.(*auth.APIKeyProvider); ok {
        apiProvider.RegisterResource(auth.ResourceConfig{
            ID:      "weather-api",
            Type:    "http",
            BaseURL: b.baseURL,
        })
    }
    
    return nil
}

func (b *WeatherBackend) Close() error {
    // Manual cleanup
    if b.authProvider != nil {
        return b.authProvider.Close()
    }
    return nil
}
```

```go
// cmd/server/main.go
import (
    "weather-mcp-server/internal/weather"
    "github.com/SaherElMasry/go-mcp-framework/backend"
    "github.com/SaherElMasry/go-mcp-framework/framework"
)

func main() {
    backend.Register("weather", func() backend.ServerBackend {
        return weather.NewWeatherBackend()
    })
    
    server := framework.NewServer(
        framework.WithBackendType("weather"),
        framework.WithConfigFile("config/config.yaml"),
    )
    
    server.Run(ctx)
}
```

---

### **After: Framework v0.3.0 Auth**

```go
// internal/weather/backend.go
import (
    "github.com/SaherElMasry/go-mcp-framework/auth" // ← Framework package
    "github.com/SaherElMasry/go-mcp-framework/backend"
)

type WeatherBackend struct {
    *backend.BaseBackend
    // No need to store authProvider - it's in BaseBackend!
    apiKey  string
}

func (b *WeatherBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
    // Setup auth using framework's system
    apiKeyProvider := auth.NewAPIKeyProvider("weather-api", auth.APIKeyConfig{
        APIKey: b.apiKey,
        Header: "key",
    })
    
    apiKeyProvider.RegisterResource(auth.ResourceConfig{
        ID:   "weather-api",
        Type: "http",
        Config: map[string]interface{}{
            "base_url": b.baseURL,
        },
    })
    
    // Use framework's method
    b.SetAuthProvider(apiKeyProvider)
    
    return nil
}

func (b *WeatherBackend) Close() error {
    // Use framework's method - handles cleanup automatically
    if provider := b.GetAuthProvider(); provider != nil {
        return provider.Close()
    }
    return nil
}
```

```go
// cmd/server/main.go
import (
    "github.com/SaherElMasry/go-mcp-framework/auth"     // ← Framework auth
    "github.com/SaherElMasry/go-mcp-framework/backend"
    "github.com/SaherElMasry/go-mcp-framework/framework"
    "github.com/SaherElMasry/go-mcp-framework/examples/weather-server/internal/weather"
)

func main() {
    backend.Register("weather", func() backend.ServerBackend {
        return weather.NewWeatherBackend()
    })
    
    // Option 1: Simple setup
    server := framework.NewServer(
        framework.WithBackendType("weather"),
        framework.WithHTTPAddress(":8080"),
        
        // Use framework auth options
        framework.WithAuth("api-key", auth.APIKeyConfig{
            APIKey: os.Getenv("WEATHER_API_KEY"),
            Header: "key",
        }),
        
        framework.WithAuthResource("default", auth.ResourceConfig{
            ID:   "weather-api",
            Type: "http",
            Config: map[string]interface{}{
                "base_url": "https://api.weatherapi.com/v1",
            },
        }),
    )
    
    // Option 2: Config file (recommended)
    server := framework.NewServer(
        framework.WithConfigFile("config/config.yaml"),
    )
    
    server.Run(ctx)
}
```

---

## 🎯 Key Benefits

| Feature | Local pkg/auth | Framework v0.3.0 Auth |
|---------|----------------|----------------------|
| **Package Location** | Your project | Framework core |
| **Setup** | Manual in each backend | Server options |
| **Lifecycle** | Manual Close() | Automatic |
| **Resources** | Manual registration | Via options |
| **Metrics** | Custom | Built-in |
| **Testing** | Custom mocks | Framework mocks |
| **Multi-provider** | Complex | `AuthManager` |
| **OAuth Support** | Would need to build | Built-in |
| **Token Storage** | Would need to build | Built-in |

---

## 📝 Step-by-Step Migration

### Step 1: Update Imports
```diff
- import "weather-mcp-server/pkg/auth"
+ import "github.com/SaherElMasry/go-mcp-framework/auth"
```

### Step 2: Update Backend
```diff
type WeatherBackend struct {
    *backend.BaseBackend
-   authProvider auth.Provider
    apiKey       string
}

func (b *WeatherBackend) Initialize(...) error {
    apiKeyProvider := auth.NewAPIKeyProvider("weather-api", auth.APIKeyConfig{
        APIKey: b.apiKey,
+       Header: "key",
    })
    
    apiKeyProvider.RegisterResource(auth.ResourceConfig{
        ID:   "weather-api",
        Type: "http",
-       BaseURL: b.baseURL,
+       Config: map[string]interface{}{"base_url": b.baseURL},
    })
    
-   b.authProvider = apiKeyProvider
+   b.SetAuthProvider(apiKeyProvider)
}

func (b *WeatherBackend) Close() error {
-   if b.authProvider != nil {
-       return b.authProvider.Close()
-   }
+   if provider := b.GetAuthProvider(); provider != nil {
+       return provider.Close()
+   }
    return nil
}
```

### Step 3: Delete Local Auth
```bash
rm -rf pkg/auth/
```

### Step 4: Update go.mod
```bash
go mod tidy
```

### Step 5: Test
```bash
export WEATHER_API_KEY=your-key
go run cmd/server/main.go
```

---

## ✅ Verification Checklist

- [ ] Imports changed to framework's `auth` package
- [ ] Backend uses `SetAuthProvider()` / `GetAuthProvider()`
- [ ] Resources use `Config: map[string]interface{}`
- [ ] Local `pkg/auth` deleted
- [ ] `go mod tidy` run successfully
- [ ] Server starts without errors
- [ ] API calls work correctly
- [ ] Metrics show auth usage

---

## 🚀 Advanced: Multiple Providers

With framework auth, you can easily use multiple providers:

```go
server := framework.NewServer()

// API Key for WeatherAPI
weatherProvider := auth.NewAPIKeyProvider("weather", auth.APIKeyConfig{
    APIKey: os.Getenv("WEATHER_API_KEY"),
    Header: "key",
})
server.GetAuthManager().Register("weather", weatherProvider)

// OAuth for GitHub (if you add GitHub integration later)
tokenStore, _ := auth.NewFileTokenStore(".tokens", os.Getenv("ENCRYPTION_KEY"))
factory := auth.NewProviderFactory(tokenStore)
githubProvider, _ := factory.Create("github", "client-id", "secret", "redirect", nil)
server.GetAuthManager().Register("github", githubProvider)

// Backend can access either
resource1, _ := backend.GetAuthenticatedResource(ctx, "weather", "weather-api")
resource2, _ := backend.GetAuthenticatedResource(ctx, "github", "github-api")
```

---

## 📚 Additional Resources

- [Framework Auth Documentation](../../auth/README.md)
- [Auth Integration Examples](../auth-examples/)
- [OAuth Provider Guide](../../auth/OAUTH.md)

---

## 💡 Tips

1. **Start with config file** - Easier to manage settings
2. **Use environment variables** - Never commit API keys
3. **Enable metrics** - Monitor auth failures
4. **Test error cases** - Invalid keys, expired tokens, etc.
5. **Add logging** - Track auth events

---

## ❓ Troubleshooting

### Error: `auth provider not found`
**Solution:** Make sure you're registering the provider before the backend initializes:
```go
framework.WithAuth("api-key", ...)  // Register first
framework.WithBackendType("weather") // Backend uses it
```

### Error: `resource not found`
**Solution:** Register resources after creating the provider:
```go
framework.WithAuth(...)
framework.WithAuthResource("default", ...) // Add resource
```

### Error: `invalid credentials`
**Solution:** Check your environment variable:
```bash
echo $WEATHER_API_KEY
```

---

**Your weather server is now using the production-ready framework auth system!** 🎉
