## Hardcover TUI

[![License: AGPL-3.0](https://img.shields.io/github/license/NotMugil/hardcover-tui)](LICENSE) [![Latest Release](https://img.shields.io/github/v/release/NotMugil/hardcover-tui)](https://github.com/NotMugil/hardcover-tui/releases/latest) [![Go](https://img.shields.io/github/go-mod/go-version/NotMugil/hardcover-tui)](go.mod)

An unofficial terminal user interface (TUI) client for [Hardcover.app](https://hardcover.app) — Social discovery for serious book lovers. Browse your library, track your reading and manage your books without leaving the terminal.

![preview](./.github/assets/preview.gif)

### Installation

#### Release Binaries
Download pre-built binaries for Linux, macOS, and Windows from the [Releases](https://github.com/NotMugil/hardcover-tui/releases) page.

#### AUR (Arch Linux)
```bash
paru -S hardcover-tui-bin
# or
yay -S hardcover-tui-bin
```

#### Go Install
```bash
go install github.com/NotMugil/hardcover-tui/cmd/hardcover-tui@latest
```

#### From Source
```bash
git clone https://github.com/NotMugil/hardcover-tui.git
cd hardcover-tui
go build -o hardcover-tui ./cmd/hardcover-tui
```

### Authentication

On first launch, `hardcover-tui` will prompt you to enter your Hardcover API key.

1. Obtain your API key from [hardcover.app/account/api](https://hardcover.app/account/api).
2. Launch the app or manage authentication via CLI:

```bash
# Save API key to system keyring
hardcover-tui auth login <your_bearer_token>

# Check current authentication status
hardcover-tui auth status

# Remove API key from keyring
hardcover-tui auth logout
```

### Usage & Keybindings

Run the TUI:
```bash
hardcover-tui
```

#### Keybindings

| Key | Action |
| --- | --- |
| `1` | Switch to **Home** (Library & Shelves) |
| `2` | Switch to **Search** |
| `3` | Switch to **Lists** |
| `4` | Switch to **Stats** & Goals |
| `Tab` / `Shift+Tab` | Cycle through tabs |
| `j` / `k` or `↓` / `↑` | Move selection down / up |
| `Enter` | Open book details / confirm selection |
| `Esc` | Go back / close screen |
| `s` | Change reading status |
| `r` | Rate or review book |
| `p` | Log reading progress |
| `?` | Toggle help overlay |
| `Ctrl+Q` | Logout |
| `q` / `Ctrl+C` | Quit |

### Contributing

Contributions are welcome! Whether it is bug fixes, new features, UI improvements, or documentation — all help is appreciated.

Please read the [Contributing Guide](./CONTRIBUTING.md) before getting started.

### Resources & Stack

- **Platform**: [Hardcover.app](https://hardcover.app) | [Hardcover API Docs](https://github.com/hardcoverapp/hardcover-docs)
- **TUI Stack**: [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), [Lip Gloss](https://github.com/charmbracelet/lipgloss), [Bubblezone](https://github.com/lrstanley/bubblezone)
- **GraphQL**: [gqlgenc](https://github.com/gqlgo/gqlgenc)
- **Keyring**: [go-keyring](https://github.com/zalando/go-keyring)

### Disclaimer

This is an independent, community-developed client and is not officially affiliated with, sponsored by, or endorsed by Hardcover. It connects to the Hardcover GraphQL API using your personal API key and acts on your behalf.

### License

This project is licensed under the [GNU Affero General Public License v3.0](./LICENSE).
