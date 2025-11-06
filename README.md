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
./json-shake [options] <json-file-path>

# Windows
json-shake.exe [options] <json-file-path>
```

### Options

- `-limit <MB>` - Maximum image size in MB (default: 0, no compression)
  - If set, images larger than the limit will be compressed to meet the size requirement
  - PNG and GIF images may be converted to JPEG for better compression

### Examples

**Download original images (no compression):**
```bash
./json-shake data.json
```

**Download with 1MB size limit:**
```bash
./json-shake -limit 1 data.json
```

**Download with 0.5MB size limit:**
```bash
./json-shake -limit 0.5 data.json
```

### Output Example

Without compression:
```
Found 18 image links
Output directory: /Users/username/Downloads/data
No size limit, downloading original images
Downloading images...
[1/18] Downloading: https://example.com/image1.png
✓ Downloaded: image1.png (0.64MB)
...
Download complete!
Success: 18, Failed: 0, Total: 18
```

With compression:
```
Found 18 image links
Output directory: /Users/username/Downloads/data
Image size limit: 1.00MB
Downloading images...
[1/18] Downloading: https://example.com/large-image.png
  Image size 5.65MB exceeds limit 1.00MB, compressing...
  Compressed from 5.65MB to 0.45MB (quality: 85)
✓ Downloaded: large-image.jpg (0.45MB)
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
- **Configurable image compression** - Set size limits to compress large images
- Intelligent quality adjustment - Automatically finds optimal compression quality
- Shows download progress with file sizes
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
