# Contributing to hardcover-tui

Thanks for your interest in contributing! This project is open to contributions of all kinds — bug fixes, new features, documentation and more.

## Getting Started

1. **Fork** the repository
2. **Clone** your fork:
   ```bash
   git clone https://github.com/<your-username>/hardcover-tui.git
   cd hardcover-tui
   ```
3. **Install dependencies**:
   ```bash
   go mod tidy
   ```
4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development

### Building

```bash
go build -o hardcover-tui ./cmd
```

### Running

```bash
go run ./cmd
```

You'll need a Hardcover API key to test. Get one from [hardcover.app/account/api](https://hardcover.app/account/api).

## Making Changes

### Code Style

- Follow standard Go conventions and formatting (`gofmt` / `goimports`).
- Keep functions focused and well-named.
- Add comments for exported types and functions.

### Commit Messages

This project follows the [Conventional Commits](https://www.conventionalcommits.org/) specification. All commit messages must follow this format:

```
<type>(<optional scope>): <description>

[optional body]

[optional footer(s)]
```

Write clear, descriptive commit messages:

```
feat(api): add book search query
fix: handle network timeout during API validation
docs: update README with installation instructions
refactor(ui): extract spinner into reusable component
```

### Before Submitting

1. Make sure the project builds without errors:
   ```bash
   go build ./...
   ```
2. Run the vet tool:
   ```bash
   go vet ./...
   ```
3. Test your changes manually by running the app.

## Submitting a Pull Request

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
2. Open a Pull Request against the `main` branch.
3. Provide a clear description of what your changes do and why.
4. Link any related issues if applicable.

## Reporting Bugs

Open an [issue](https://github.com/NotMugil/hardcover-tui/issues) with:

- A clear title and description
- Steps to reproduce the problem
- Expected vs actual behavior
- Your OS and Go version
- Clear Screenshots or video of the bug, if applicable

## Suggesting Features

Feature ideas are welcome! Open an [issue](https://github.com/NotMugil/hardcover-tui/issues) and describe:

- What you'd like to see
- Why it would be useful
- Any implementation ideas you have
## License

By contributing, you agree that your contributions will be licensed under the [AGPL-3.0 License](LICENSE).
