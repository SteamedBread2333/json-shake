# JSON Shake

A command-line tool to recursively extract and download all image URLs from JSON files. Supports macOS and Windows.

## Installation

No installation required. Use the pre-compiled binaries:
- **macOS**: `json-shake`
- **Windows**: `json-shake.exe`

## Usage

### Basic Command

```bash
# macOS/Linux
./json-shake <json-file-path>

# Windows
json-shake.exe <json-file-path>
```

### Example

```bash
./json-shake data.json
```

Output:
```
Found 18 image links
Output directory: /Users/username/Downloads/data
Downloading images...
[1/18] Downloading: https://example.com/image1.png
✓ Downloaded: image1.png
[2/18] Downloading: https://example.com/image2.jpg
✓ Downloaded: image2.jpg
...
Download complete!
Success: 18, Failed: 0, Total: 18
```

### Output Location

Images are downloaded to:
- **macOS**: `~/Downloads/<json-filename>/`
- **Windows**: `C:\Users\<username>\Downloads\<json-filename>\`

## Supported Image Formats

- JPG/JPEG
- PNG
- GIF
- BMP
- WebP
- SVG

## Features

- Recursively parses nested JSON structures
- Automatically detects image URLs (with or without file extensions)
- Batch downloads all images
- Shows download progress
- Skips already downloaded files
- Cross-platform support (macOS/Windows)

## Building from Source

Requires Go 1.21 or higher:

```bash
# Build for current platform
go build -o json-shake main.go

# Cross-compile for Windows (from macOS/Linux)
GOOS=windows GOARCH=amd64 go build -o json-shake.exe main.go

# Cross-compile for macOS (from Windows)
set GOOS=darwin
set GOARCH=amd64
go build -o json-shake main.go
```

## License

MIT License
