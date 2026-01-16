
# Contributing to go-mcp-framework

First off, thank you for considering contributing to go-mcp-framework! It's people like you that make this framework better for everyone.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct: be respectful, collaborative, and constructive.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* Use a clear and descriptive title
* Describe the exact steps which reproduce the problem
* Provide specific examples to demonstrate the steps
* Describe the behavior you observed and what behavior you expected
* Include logs and error messages

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* Use a clear and descriptive title
* Provide a step-by-step description of the suggested enhancement
* Provide specific examples to demonstrate the steps
* Describe the current behavior and explain the behavior you expected to see
* Explain why this enhancement would be useful

### Pull Requests

* Fill in the required template
* Follow the Go style guide
* Include appropriate test coverage
* Update documentation as needed
* End all files with a newline

## Development Process

1. Fork the repo
2. Create a new branch from `main`
3. Make your changes
4. Write or update tests
5. Ensure all tests pass
6. Submit a pull request

### Development Setup
```bash
git clone https://github.com/SaherElMasry/go-mcp-framework.git
cd go-mcp-framework
go mod download
```

### Running Tests
```bash
go test ./...
```

### Code Style

* Follow standard Go conventions
* Run `gofmt` before committing
* Use meaningful variable and function names
* Comment exported functions and types
* Keep functions focused and small

## Project Structure
```
go-mcp-framework/
├── backend/         # Backend interface and implementations
├── framework/       # Core framework
├── protocol/        # JSON-RPC and MCP protocol
├── transport/       # Communication layers
├── observability/   # Metrics and logging
└── examples/        # Example implementations
```

## Questions?

Feel free to open an issue with your question or reach out via discussions!

Thank you! 🎉
