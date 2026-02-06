# screenshot

Fast, flexible screenshot utility for Linux.

## Features

- Capture all monitors or specific monitor
- Region capture
- Multiple compression levels
- Output to file or stdout (for piping)
- Open in default viewer
- Works when screen is locked (via cron with `-d :0`)
- Strategy-based architecture (X11 now, Wayland/Windows/macOS ready)

## Installation

```bash
go install github.com/robotin/screenshot@latest
```

Or build from source:

```bash
git clone https://github.com/robotin/screenshot.git
cd screenshot
go build -o bin/screenshot .
sudo ln -sf $(pwd)/bin/screenshot /usr/bin/screenshot
```

## Usage

```bash
screenshot                      # Capture all monitors, fast compression
screenshot captura.png          # Capture to specific file
screenshot -r                   # No compression (raw, fastest)
screenshot -ccc                 # Best compression (smallest)
screenshot -v                   # Capture and open in viewer
screenshot --stdout | feh -     # Pipe to image viewer
screenshot -m 0                 # Capture only monitor 0
screenshot -m 1                 # Capture only monitor 1
screenshot --region 100,100,500,400   # Capture region (x,y,width,height)
screenshot -d :0                # Force DISPLAY (for cron)
screenshot --list               # List available monitors
```

## Compression Levels

| Flag | Level | Speed | Size |
|------|-------|-------|------|
| `-r` | raw | ~0.1s | largest |
| (default) | fast | ~0.3s | medium |
| `-cc` | medium | ~1.2s | smaller |
| `-ccc` | best | ~9s | smallest |

## License

MIT
