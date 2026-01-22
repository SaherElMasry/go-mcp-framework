// framework/color_helper.go
package framework

import (
	"fmt"
	"os"

	"github.com/SaherElMasry/go-mcp-framework/color"
)

// PrintStartupBanner prints a colorful startup banner
func PrintStartupBanner(name, version, description string) {
	fmt.Println(color.Banner(
		fmt.Sprintf("%s v%s", name, version),
		description,
	))
}

// PrintServerConfig prints server configuration in a colorful table
func PrintServerConfig(config map[string]string) {
	table := color.NewTable("Setting", "Value")

	for key, value := range config {
		table.AddRow(key, value)
	}

	fmt.Println(color.Info("Server Configuration:"))
	fmt.Println(table.String())
	fmt.Println()
}

// PrintStartupMessage prints a formatted startup message
func PrintStartupMessage(transport, address string) {
	fmt.Println(color.Success("Server starting"))
	fmt.Printf("  %s %s\n", color.Cyan("Transport:"), color.MakeBold(transport))
	fmt.Printf("  %s %s\n", color.Cyan("Address:"), color.MakeBold(address))
	fmt.Println()
}

// PrintShutdownMessage prints a shutdown message
func PrintShutdownMessage() {
	fmt.Println()
	fmt.Println(color.Warning("Shutting down gracefully..."))
}

// PrintError prints a colored error message
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "%s %v\n", color.Error("Error:"), err)
}

// PrintToolsList prints available tools in a colored format
func PrintToolsList(tools []string) {
	fmt.Println(color.Info("Available Tools:"))
	for i, tool := range tools {
		// Fix: Format the number first, then colorize
		num := fmt.Sprintf("%d.", i+1)
		fmt.Printf("  %s %s\n",
			color.Gray("%s", num),
			color.Cyan("%s", tool))
	}
	fmt.Println()
}

// PrintMetrics prints metrics in a colored format
func PrintMetrics(metrics map[string]interface{}) {
	fmt.Println(color.Info("Metrics:"))

	for key, value := range metrics {
		var coloredValue string
		switch v := value.(type) {
		case int, int64:
			coloredValue = color.Magenta("%v", v)
		case float64:
			coloredValue = color.Magenta("%.2f", v)
		case string:
			// Fix: Don't use variable as format string
			coloredValue = color.Green("%s", v)
		case bool:
			if v {
				coloredValue = color.BrightGreen("true")
			} else {
				coloredValue = color.BrightRed("false")
			}
		default:
			coloredValue = fmt.Sprintf("%v", v)
		}

		// Fix: Format the key first, then colorize
		keyStr := key + ":"
		keyColored := color.Cyan("%s", keyStr)
		fmt.Printf("  %s %s\n", keyColored, coloredValue)
	}
	fmt.Println()
}

// WithColors returns a function option to enable colored output
func WithColors(enable bool) Option {
	return func(s *Server) {
		if enable {
			color.Enable()
		} else {
			color.Disable()
		}
	}
}

// WithAutoColors returns a function option to auto-detect color support
func WithAutoColors() Option {
	return func(s *Server) {
		color.AutoDetect()
	}
}
