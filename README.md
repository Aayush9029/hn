# hn

Browse Hacker News from the terminal — CLI + interactive TUI.

## Installation

```bash
brew install aayush9029/tap/hn
```

Or tap first:

```bash
brew tap aayush9029/tap
brew install hn
```

## Usage

```bash
hn top 10              # top 10 stories
hn new 20              # 20 newest stories
hn ask                 # Ask HN discussions (default 30)
hn thread 12345678     # read a thread with comments
hn search "rust lang"  # search stories via Algolia
hn -i                  # interactive TUI
```

## Options

| Flag | Description |
|------|-------------|
| `top/new/best/ask/show/jobs [n]` | List stories (default 30) |
| `thread, t <id>` | View story + nested comments |
| `search, s <query> [n]` | Search stories (default 20) |
| `-i, --interactive` | Launch full TUI |
| `-v, --version` | Print version |
| `-h, --help` | Show help |

## How it works

1. Fetches story IDs from the HN Firebase API
2. Concurrently resolves individual items (20 workers)
3. Recursively builds comment trees with depth limiting
4. Renders in the terminal with ANSI colors or BubbleTea TUI

## Requirements

- macOS (or Linux)

## License

MIT
