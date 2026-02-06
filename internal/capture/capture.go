package capture

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/robotin/screenshot/internal/strategy"
)

// Capturer handles screenshot capture with strategy selection
type Capturer struct {
	strategies []strategy.Strategy
}

// New creates a new Capturer with available strategies
func New() *Capturer {
	c := &Capturer{
		strategies: []strategy.Strategy{},
	}

	// Add X11 strategy for Linux
	x11 := strategy.NewX11Strategy()
	if x11.Available() {
		c.strategies = append(c.strategies, x11)
	}

	// TODO: Add Wayland strategy
	// TODO: Add Windows strategy
	// TODO: Add macOS strategy

	return c
}

// GetStrategy returns the first available strategy
func (c *Capturer) GetStrategy() (strategy.Strategy, error) {
	if len(c.strategies) == 0 {
		return nil, fmt.Errorf("no screenshot strategy available")
	}
	return c.strategies[0], nil
}

// ListStrategies returns all available strategy names
func (c *Capturer) ListStrategies() []string {
	names := make([]string, len(c.strategies))
	for i, s := range c.strategies {
		names[i] = s.Name()
	}
	return names
}

// CaptureToFile captures a screenshot and saves it to a file
// compressionLevel: 0=None, 1=BestSpeed, 2=Default, 3=BestCompression
func (c *Capturer) CaptureToFile(opts strategy.CaptureOptions, outputPath string, compressionLevel int) error {
	strat, err := c.GetStrategy()
	if err != nil {
		return err
	}

	img, err := strat.Capture(opts)
	if err != nil {
		return fmt.Errorf("capture failed: %w", err)
	}

	return SavePNG(img, outputPath, compressionLevel)
}

// Capture captures a screenshot and returns the image
func (c *Capturer) Capture(opts strategy.CaptureOptions) (image.Image, error) {
	strat, err := c.GetStrategy()
	if err != nil {
		return nil, err
	}

	return strat.Capture(opts)
}

// ListMonitors returns available monitors
func (c *Capturer) ListMonitors() ([]strategy.Monitor, error) {
	strat, err := c.GetStrategy()
	if err != nil {
		return nil, err
	}

	return strat.ListMonitors()
}

// SavePNG saves an image to a PNG file
// compressionLevel: 0=None, 1=BestSpeed, 2=Default, 3=BestCompression
func SavePNG(img image.Image, path string, compressionLevel int) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := png.Encoder{CompressionLevel: intToCompressionLevel(compressionLevel)}
	if err := encoder.Encode(file, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// GenerateFilename generates a default filename with timestamp
func GenerateFilename(prefix string) string {
	if prefix == "" {
		prefix = "screenshot"
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("%s_%s.png", prefix, timestamp)
}

// WritePNG writes an image as PNG to any io.Writer
// compressionLevel: 0=None, 1=BestSpeed, 2=Default, 3=BestCompression
func WritePNG(img image.Image, w io.Writer, compressionLevel int) error {
	encoder := png.Encoder{CompressionLevel: intToCompressionLevel(compressionLevel)}
	if err := encoder.Encode(w, img); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	return nil
}

// intToCompressionLevel converts int to png.CompressionLevel
func intToCompressionLevel(level int) png.CompressionLevel {
	switch level {
	case 0:
		return png.NoCompression
	case 1:
		return png.BestSpeed
	case 2:
		return png.DefaultCompression
	case 3:
		return png.BestCompression
	default:
		return png.BestSpeed
	}
}
