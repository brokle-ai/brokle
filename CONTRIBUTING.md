# Contributing to Brokle

Thank you for your interest in contributing! We welcome contributions from the community.

## Getting Started

### Setup
```bash
git clone https://github.com/yourusername/brokle-platform.git
cd brokle-platform/brokle
make setup
make dev
```

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose

## How to Contribute

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/your-feature`
3. Make your changes
4. Write tests
5. Run tests: `make test`
6. Submit a pull request

## Code Standards

- Follow Go conventions
- Use meaningful variable names
- Write tests for new functionality
- Update documentation when needed
- Run `make lint` before submitting

## Development Commands

```bash
make setup          # Initial setup
make dev           # Start development
make test          # Run tests
make lint          # Code linting
make fmt           # Code formatting
```

## Reporting Issues

When reporting bugs, include:
- Description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment details
- Relevant logs

## License

By contributing, you agree that your contributions will be licensed under the MIT License.