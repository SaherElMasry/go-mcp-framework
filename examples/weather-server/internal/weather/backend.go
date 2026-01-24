// examples/weather-server/internal/weather/backend.go
package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/SaherElMasry/go-mcp-framework/auth"
	"github.com/SaherElMasry/go-mcp-framework/backend"
)

// WeatherBackend implements MCP server for weather data
type WeatherBackend struct {
	*backend.BaseBackend
	apiKey  string
	baseURL string
	timeout time.Duration
	cache   map[string]*CachedWeather
}

// CachedWeather stores cached weather data
type CachedWeather struct {
	Data      interface{}
	ExpiresAt time.Time
}

type AstronomyData struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		Time    string `json:"localtime"`
	} `json:"location"`
	Astronomy struct {
		Astro struct {
			Sunrise      string  `json:"sunrise"`
			Sunset       string  `json:"sunset"`
			Moonrise     string  `json:"moonrise"`
			Moonset      string  `json:"moonset"`
			MoonPhase    string  `json:"moon_phase"`
			Illumination float64 `json:"moon_illumination"`
		} `json:"astro"`
	} `json:"astronomy"`
}

// NewWeatherBackend creates a new weather backend
func NewWeatherBackend() *WeatherBackend {
	b := &WeatherBackend{
		BaseBackend: backend.NewBaseBackend("Weather API"),
		cache:       make(map[string]*CachedWeather),
		timeout:     30 * time.Second,
		baseURL:     "https://api.weatherapi.com/v1",
	}

	// Register all tools
	b.registerTools()

	return b
}

// Initialize sets up the backend with configuration
func (b *WeatherBackend) Initialize(ctx context.Context, config map[string]interface{}) error {
	// Try to get API key from config (from YAML or programmatic config)
	if apiKey, ok := config["api_key"].(string); ok && apiKey != "" {
		b.apiKey = apiKey
	}

	// If not in config, try to get from auth provider (set by framework)
	if b.apiKey == "" {
		if provider := b.GetAuthProvider(); provider != nil {
			// Auth provider was set by framework - extract the API key
			if apiKeyProvider, ok := provider.(*auth.APIKeyProvider); ok {
				// The APIKeyProvider doesn't expose the key directly for security,
				// but we can validate it which ensures it's set
				if err := apiKeyProvider.Validate(ctx); err == nil {
					// Key is valid, but we need to get it somehow
					// For now, we'll use a workaround - the framework should pass it in config
					b.apiKey = "from-auth-provider"
				}
			}
		}
	}

	// Parse other configuration
	if baseURL, ok := config["base_url"].(string); ok {
		b.baseURL = baseURL
	}
	if timeout, ok := config["timeout"].(time.Duration); ok {
		b.timeout = timeout
	}

	// Final validation - we MUST have an API key at this point
	if b.apiKey == "" {
		return fmt.Errorf("missing API key in configuration - set WEATHER_API_KEY environment variable")
	}

	return nil
}

// registerTools registers all available tools
// In registerTools() function, update tool definitions:
func (b *WeatherBackend) registerTools() {
	// 1. Current Weather - CACHEABLE (5 minutes)
	b.RegisterTool(
		backend.NewTool("get_current_weather").
			Description("Get current weather for a location").
			StringParam("location", "City name, zip code, or coordinates", true).
			WithCache(true, 5*time.Minute). // 🆕 CACHEABLE!
			Build(),
		b.handleGetCurrentWeather,
	)

	// 2. Weather Forecast - CACHEABLE (30 minutes)
	b.RegisterTool(
		backend.NewTool("get_forecast").
			Description("Get weather forecast for up to 10 days").
			StringParam("location", "City name, zip code, or coordinates", true).
			IntParam("days", "Number of forecast days (1-10)", false, nil, intPtr(10)).
			WithCache(true, 30*time.Minute). // 🆕 CACHEABLE!
			Build(),
		b.handleGetForecast,
	)

	// 3. Search Locations - STREAMING (not cacheable)
	b.RegisterStreamingTool(
		backend.NewTool("search_locations").
			Description("Search for locations with real-time results").
			StringParam("query", "Search query", true).
			NonCacheable(). // 🆕 Explicitly non-cacheable
			Streaming(true).
			Build(),
		b.handleSearchLocations,
	)

	// 4. Astronomy - CACHEABLE (1 hour)
	b.RegisterTool(
		backend.NewTool("get_astronomy").
			Description("Get astronomy data (sunrise, sunset, moon phase)").
			StringParam("location", "City name, zip code, or coordinates", true).
			StringParam("date", "Date in YYYY-MM-DD format (optional)", false).
			WithCache(true, 1*time.Hour). // 🆕 CACHEABLE!
			Build(),
		b.handleGetAstronomy,
	)

	// 5. Bulk Check - STREAMING (not cacheable)
	b.RegisterStreamingTool(
		backend.NewTool("bulk_weather_check").
			Description("Get weather for multiple locations").
			StringParam("locations", "Comma-separated list", true).
			NonCacheable(). // 🆕 Explicitly non-cacheable
			Streaming(true).
			Build(),
		b.handleBulkWeatherCheck,
	)
}

// handleGetCurrentWeather gets current weather for a location
func (b *WeatherBackend) handleGetCurrentWeather(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	location := args["location"].(string)

	// Check cache first
	cacheKey := fmt.Sprintf("current:%s", location)
	if cached, exists := b.cache[cacheKey]; exists && time.Now().Before(cached.ExpiresAt) {
		return cached.Data, nil
	}

	// Build API URL
	apiURL := b.buildURL("/current.json", map[string]string{
		"q":   location,
		"aqi": "no",
	})

	// Make API request
	resp, err := b.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if errMsg, exists := result["error"]; exists {
		return nil, fmt.Errorf("API error: %v", errMsg)
	}

	// Extract nested data
	loc, _ := result["location"].(map[string]interface{})
	curr, _ := result["current"].(map[string]interface{})
	cond, _ := curr["condition"].(map[string]interface{})

	// Create readable summary
	readableSummary := fmt.Sprintf(
		"Current weather in %s, %s (%s):\n"+
			"🌡️ Temperature: %.1f°C (Feels like %.1f°C)\n"+
			"☁️ Condition: %s\n"+
			"💧 Humidity: %.0f%%\n"+
			"🌬️ Wind: %.1f kph coming from %s\n"+
			"🕒 Last Updated: %s",
		loc["name"], loc["region"], loc["country"],
		curr["temp_c"], curr["feelslike_c"],
		cond["text"],
		curr["humidity"],
		curr["wind_kph"], curr["wind_dir"],
		curr["last_updated"],
	)

	finalOutput := map[string]interface{}{
		"summary":  readableSummary,
		"raw_data": result,
	}

	// Cache the output
	b.cache[cacheKey] = &CachedWeather{
		Data:      finalOutput,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	return finalOutput, nil
}

// handleGetForecast gets weather forecast
func (b *WeatherBackend) handleGetForecast(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	location := args["location"].(string)
	days := 3
	if d, ok := args["days"].(float64); ok {
		days = int(d)
	}

	apiURL := b.buildURL("/forecast.json", map[string]string{
		"q":    location,
		"days": fmt.Sprintf("%d", days),
		"aqi":  "no",
	})

	resp, err := b.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if errMsg, exists := result["error"]; exists {
		return nil, fmt.Errorf("API error: %v", errMsg)
	}

	// Create readable summary
	forecast := result["forecast"].(map[string]interface{})
	forecastDays := forecast["forecastday"].([]interface{})
	loc := result["location"].(map[string]interface{})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("📅 Forecast for %s, %s:\n", loc["name"], loc["country"]))

	for _, day := range forecastDays {
		d := day.(map[string]interface{})
		date := d["date"].(string)
		dayData := d["day"].(map[string]interface{})
		cond := dayData["condition"].(map[string]interface{})

		sb.WriteString(fmt.Sprintf("\n• %s:\n", date))
		sb.WriteString(fmt.Sprintf("  🌡️ Max: %.1f°C, Min: %.1f°C\n", dayData["maxtemp_c"], dayData["mintemp_c"]))
		sb.WriteString(fmt.Sprintf("  ☁️ Condition: %s\n", cond["text"]))
		sb.WriteString(fmt.Sprintf("  💧 Rain Chance: %v%%\n", dayData["daily_chance_of_rain"]))
	}

	return map[string]interface{}{
		"summary":  sb.String(),
		"raw_data": result,
	}, nil
}

// handleSearchLocations searches for locations (STREAMING)
func (b *WeatherBackend) handleSearchLocations(
	ctx context.Context,
	args map[string]interface{},
	emit backend.StreamingEmitter,
) error {
	query := args["query"].(string)

	emit.EmitProgress(0, 100, "Starting location search...")

	apiURL := b.buildURL("/search.json", map[string]string{
		"q": query,
	})

	emit.EmitProgress(30, 100, "Querying WeatherAPI...")

	resp, err := b.makeRequest(ctx, apiURL)
	if err != nil {
		return fmt.Errorf("API request failed: %w", err)
	}

	emit.EmitProgress(60, 100, "Processing results...")

	var locations []map[string]interface{}
	if err := json.Unmarshal(resp, &locations); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	for i, location := range locations {
		select {
		case <-emit.Context().Done():
			return ctx.Err()
		default:
		}

		emit.EmitData(map[string]interface{}{
			"index":    i + 1,
			"total":    len(locations),
			"location": location,
		})

		progress := 60 + (40 * (i + 1) / len(locations))
		emit.EmitProgress(int64(progress), 100,
			fmt.Sprintf("Processed %d/%d locations", i+1, len(locations)))
	}

	return nil
}

// handleGetAstronomy gets astronomy data
func (b *WeatherBackend) handleGetAstronomy(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	location := args["location"].(string)
	date := time.Now().Format("2006-01-02")
	if d, ok := args["date"].(string); ok && d != "" {
		date = d
	}

	apiURL := b.buildURL("/astronomy.json", map[string]string{
		"q":  location,
		"dt": date,
	})

	resp, err := b.makeRequest(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	var result AstronomyData
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	summary := fmt.Sprintf(
		"🔭 Astronomy Report for %s, %s (%s):\n"+
			"☀️ Sun: Sunrise at %s, Sunset at %s\n"+
			"🌙 Moon: %s Phase with %.0f%% illumination. Rises at %s, Sets at %s",
		result.Location.Name, result.Location.Country, date,
		result.Astronomy.Astro.Sunrise, result.Astronomy.Astro.Sunset,
		result.Astronomy.Astro.MoonPhase, result.Astronomy.Astro.Illumination,
		result.Astronomy.Astro.Moonrise, result.Astronomy.Astro.Moonset,
	)

	return summary, nil
}

// handleBulkWeatherCheck gets weather for multiple locations (STREAMING)
func (b *WeatherBackend) handleBulkWeatherCheck(
	ctx context.Context,
	args map[string]interface{},
	emit backend.StreamingEmitter,
) error {
	locationsStr := args["locations"].(string)

	locations := parseLocations(locationsStr)
	total := len(locations)

	emit.EmitProgress(0, int64(total), fmt.Sprintf("Starting bulk check for %d locations", total))

	for i, location := range locations {
		select {
		case <-emit.Context().Done():
			return ctx.Err()
		default:
		}

		weather, err := b.getCurrentWeatherData(ctx, location)

		result := map[string]interface{}{
			"index":    i + 1,
			"location": location,
		}

		if err != nil {
			result["error"] = err.Error()
			result["success"] = false
		} else {
			result["weather"] = weather
			result["success"] = true
		}

		emit.EmitData(result)
		emit.EmitProgress(int64(i+1), int64(total),
			fmt.Sprintf("Processed %d/%d locations", i+1, total))

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// buildURL builds API URL with authentication
func (b *WeatherBackend) buildURL(endpoint string, params map[string]string) string {
	u, _ := url.Parse(b.baseURL)
	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(endpoint, "/")

	q := u.Query()
	q.Set("key", b.apiKey)
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// makeRequest makes an HTTP request with timeout
func (b *WeatherBackend) makeRequest(ctx context.Context, url string) ([]byte, error) {
	client := &http.Client{Timeout: b.timeout}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// getCurrentWeatherData is a helper
func (b *WeatherBackend) getCurrentWeatherData(ctx context.Context, location string) (map[string]interface{}, error) {
	args := map[string]interface{}{"location": location}
	result, err := b.handleGetCurrentWeather(ctx, args)
	if err != nil {
		return nil, err
	}
	return result.(map[string]interface{}), nil
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func parseLocations(str string) []string {
	parts := strings.Split(str, ",")
	locations := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			locations = append(locations, trimmed)
		}
	}
	return locations
}

// Close cleans up resources
func (b *WeatherBackend) Close() error {
	b.cache = make(map[string]*CachedWeather)

	// Close auth provider using framework's method
	if provider := b.GetAuthProvider(); provider != nil {
		return provider.Close()
	}

	return nil
}
