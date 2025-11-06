package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "image/gif"
	_ "image/png"
)

// Regular expression pattern for image URLs
var imageURLPattern = regexp.MustCompile(`https?://[^\s"'<>]+\.(?:jpg|jpeg|png|gif|bmp|webp|svg)(?:\?[^\s"'<>]*)?`)

// Check if a string is possibly an image URL (including URLs without explicit extensions)
func isPossibleImageURL(s string) bool {
	// First try to match explicit image extensions
	if imageURLPattern.MatchString(s) {
		return true
	}

	// Check if it's an HTTP/HTTPS URL
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return false
	}

	// Try to parse the URL
	_, err := url.Parse(s)
	if err != nil {
		return false
	}

	// Check if URL contains common image-related keywords
	lowerURL := strings.ToLower(s)
	imageKeywords := []string{"image", "img", "photo", "picture", "pic", "avatar", "thumbnail", "thumb", "banner", "gallery"}
	for _, keyword := range imageKeywords {
		if strings.Contains(lowerURL, keyword) {
			return true
		}
	}

	return false
}

// Get file extension from Content-Type
func getExtensionFromContentType(contentType string) string {
	contentType = strings.ToLower(strings.Split(contentType, ";")[0])
	contentType = strings.TrimSpace(contentType)

	extensions := map[string]string{
		"image/jpeg":    ".jpg",
		"image/jpg":     ".jpg",
		"image/png":     ".png",
		"image/gif":     ".gif",
		"image/bmp":     ".bmp",
		"image/webp":    ".webp",
		"image/svg+xml": ".svg",
	}

	if ext, ok := extensions[contentType]; ok {
		return ext
	}
	return ""
}

// Recursively traverse JSON object and extract all image links
func extractImageURLs(data interface{}, urls *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Traverse JSON object
		for _, value := range v {
			extractImageURLs(value, urls)
		}
	case []interface{}:
		// Traverse JSON array
		for _, item := range v {
			extractImageURLs(item, urls)
		}
	case string:
		// First check if string contains explicit image URLs
		matches := imageURLPattern.FindAllString(v, -1)
		*urls = append(*urls, matches...)

		// If no explicit image URLs found, check if it's possibly an image URL
		if len(matches) == 0 && isPossibleImageURL(v) {
			*urls = append(*urls, v)
		}
	}
}

// Compress image if it exceeds the size limit
func compressImage(data []byte, limitMB float64) ([]byte, error) {
	limitBytes := int64(limitMB * 1024 * 1024)

	// If image is within limit, return original
	if int64(len(data)) <= limitBytes {
		return data, nil
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Try different quality levels to meet the size limit
	qualities := []int{85, 75, 65, 55, 45, 35, 25}

	for _, quality := range qualities {
		var buf bytes.Buffer

		switch format {
		case "jpeg", "jpg":
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		case "png":
			// PNG compression is lossless, so we convert to JPEG for lossy compression
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		case "gif":
			// GIF compression - just return original or convert to JPEG
			err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		default:
			return data, nil // Return original for unsupported formats
		}

		if err != nil {
			continue
		}

		// Check if compressed size is within limit
		if int64(buf.Len()) <= limitBytes {
			fmt.Printf("  Compressed from %.2fMB to %.2fMB (quality: %d)\n",
				float64(len(data))/1024/1024,
				float64(buf.Len())/1024/1024,
				quality)
			return buf.Bytes(), nil
		}
	}

	// If still too large, return the most compressed version
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 20})
	fmt.Printf("  Compressed from %.2fMB to %.2fMB (quality: 20 - minimum)\n",
		float64(len(data))/1024/1024,
		float64(buf.Len())/1024/1024)
	return buf.Bytes(), nil
}

// Download image to specified directory
func downloadImage(imageURL, outputDir string, index int, limitMB float64) error {
	// Parse URL
	parsedURL, err := url.Parse(imageURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// Get filename
	filename := filepath.Base(parsedURL.Path)
	if filename == "" || filename == "." || filename == "/" {
		filename = fmt.Sprintf("image_%d", index)
	}

	// Clean special characters in filename
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "&", "_")

	// If filename has no extension, try to get it from Content-Type
	if !strings.Contains(filename, ".") {
		filename = fmt.Sprintf("%s_%d", filename, index)
	}

	// Build full output path
	outputPath := filepath.Join(outputDir, filename)

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		fmt.Printf("File already exists, skipping: %s\n", filename)
		return nil
	}

	// Send HTTP request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(imageURL)
	if err != nil {
		return fmt.Errorf("download failed: %v", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP error: %s", resp.Status)
	}

	// If filename has no extension, try to infer from Content-Type
	if !strings.Contains(filename, ".") {
		contentType := resp.Header.Get("Content-Type")
		ext := getExtensionFromContentType(contentType)
		if ext != "" {
			filename = filename + ext
			outputPath = filepath.Join(outputDir, filename)
		}
	}

	// Read image data into memory
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Apply compression if limit is set
	if limitMB > 0 {
		originalSize := float64(len(imageData)) / 1024 / 1024
		if originalSize > limitMB {
			fmt.Printf("  Image size %.2fMB exceeds limit %.2fMB, compressing...\n", originalSize, limitMB)
			ext := filepath.Ext(filename)
			imageData, err = compressImage(imageData, limitMB)
			if err != nil {
				fmt.Printf("  Warning: compression failed, saving original: %v\n", err)
			} else {
				// Update filename extension if changed during compression
				if ext == ".png" || ext == ".gif" {
					filename = strings.TrimSuffix(filename, ext) + ".jpg"
					outputPath = filepath.Join(outputDir, filename)
				}
			}
		}
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	// Write to file
	_, err = outFile.Write(imageData)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	finalSize := float64(len(imageData)) / 1024 / 1024
	fmt.Printf("✓ Downloaded: %s (%.2fMB)\n", filename, finalSize)
	return nil
}

// Get user's Download directory
func getDownloadDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Cross-platform Download directory
	downloadDir := filepath.Join(homeDir, "Downloads")
	return downloadDir, nil
}

func main() {
	// Define command line flags
	var limitMB float64
	flag.Float64Var(&limitMB, "limit", 0, "Maximum image size in MB (0 = no limit, download original)")
	flag.Parse()

	// Check command line arguments
	if flag.NArg() < 1 {
		fmt.Println("Usage: json-shake [options] <json-file-path>")
		fmt.Println("Options:")
		fmt.Println("  -limit <MB>  Maximum image size in MB (default: 0, no compression)")
		fmt.Println("Example: json-shake data.json")
		fmt.Println("Example: json-shake -limit 1 data.json")
		os.Exit(1)
	}

	jsonFilePath := flag.Arg(0)

	// Read JSON file
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		fmt.Printf("Failed to read file: %v\n", err)
		os.Exit(1)
	}

	// Parse JSON
	var data interface{}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	// Extract all image URLs
	var imageURLs []string
	extractImageURLs(data, &imageURLs)

	if len(imageURLs) == 0 {
		fmt.Println("No image links found")
		os.Exit(0)
	}

	fmt.Printf("Found %d image links\n", len(imageURLs))

	// Get JSON filename (without extension)
	jsonFileName := strings.TrimSuffix(filepath.Base(jsonFilePath), filepath.Ext(jsonFilePath))

	// Get Download directory
	downloadDir, err := getDownloadDir()
	if err != nil {
		fmt.Printf("Failed to get Download directory: %v\n", err)
		os.Exit(1)
	}

	// Create output directory
	outputDir := filepath.Join(downloadDir, jsonFileName)
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Output directory: %s\n", outputDir)
	if limitMB > 0 {
		fmt.Printf("Image size limit: %.2fMB\n", limitMB)
	} else {
		fmt.Println("No size limit, downloading original images")
	}
	fmt.Println("Downloading images...")

	// Download all images
	successCount := 0
	failCount := 0
	for i, imageURL := range imageURLs {
		fmt.Printf("[%d/%d] Downloading: %s\n", i+1, len(imageURLs), imageURL)
		err := downloadImage(imageURL, outputDir, i+1, limitMB)
		if err != nil {
			fmt.Printf("✗ Error: %v\n", err)
			failCount++
		} else {
			successCount++
		}
	}

	// Output statistics
	fmt.Println("\nDownload complete!")
	fmt.Printf("Success: %d, Failed: %d, Total: %d\n", successCount, failCount, len(imageURLs))
}
