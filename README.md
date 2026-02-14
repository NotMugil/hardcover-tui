## Hardcover TUI

[![License: AGPL-3.0](https://img.shields.io/github/license/NotMugil/hardcover-tui)](LICENSE) [![Latest Release](https://img.shields.io/github/v/release/NotMugil/hardcover-tui)](https://github.com/NotMugil/hardcover-tui/releases/latest) [![Go](https://img.shields.io/github/go-mod/go-version/NotMugil/hardcover-tui)](go.mod)

An unofficial terminal user interface (TUI) client for [Hardcover.app](https://hardcover.app) — Social discovery for serious book lovers. Browse your library, track your reading and manage your books without leaving the terminal.


![preview](./assets/preview.gif)

### Installation

#### From Release Binaries

Download the latest binary for your platform from the [Releases](https://github.com/NotMugil/hardcover-tui/releases) page.

#### From Source

```bash
git clone https://github.com/NotMugil/hardcover-tui.git
cd hardcover-tui
go mod tidy
go build -o hardcover-tui ./cmd
```

#### Go Install


```bash
go install github.com/NotMugil/hardcover-tui/cmd/hardcover-tui@latest
```

### Usage

```bash
./hardcover-tui
```


On first launch, you'll be prompted to enter your Hardcover API key. Visit [hardcover.app/account/api](https://hardcover.app/account/api) to get your API key and then copy and paste it into the app when prompted.

### Contributing
Contributions are welcome! Whether it is opening an issue, bug fixes, new features, documentation improvements or document translations — all help is appreciated.

Please read the [Contributing Guide](./CONTRIBUTING.md) before getting started.

### Resources

- [Hardcover Website](https://hardcover.app)  
- [Hardcover API Docs](https://github.com/hardcoverapp/hardcover-docs)
- [Kameleon21/oku](https://github.com/Kameleon21/oku)

#### Related Libraries
- [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- [go-graphql-client](https://github.com/hasura/go-graphql-client)
- [go-keyring](https://github.com/zalando/go-keyring)

### Disclaimer
This is an independent, community-developed client and is not affiliated with, sponsored by, or endorsed by Hardcover. It connects to the Hardcover GraphQL API using your personal API key and acts on your behalf.

### License
This project is licensed under the [GNU Affero General Public License v3.0.](./LICENSE)
