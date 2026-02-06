package cmd

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/robotin/screenshot/internal/capture"
	"github.com/robotin/screenshot/internal/strategy"
	"github.com/spf13/cobra"
)

var (
	// Flags
	monitor       int
	region        string
	output        string
	display       string
	listMon       bool
	compressLevel int
	raw           bool
	view          bool
	stdout        bool
)

var rootCmd = &cobra.Command{
	Use:   "screenshot [output]",
	Short: "Fast screenshot utility",
	Long: `A fast, flexible screenshot utility for Linux.

Captures screenshots with support for multiple monitors, regions,
and works even when the screen is locked (via cron).

Compression levels:
  -r            Raw, no compression (fastest, largest)
  -c            Fast compression (default)
  -cc           Medium compression
  -ccc          Best compression (slowest, smallest)

Examples:
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
  screenshot --list               # List available monitors`,
	Args: cobra.MaximumNArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().IntVarP(&monitor, "monitor", "m", -1, "Monitor index to capture (-1 = all, default)")
	rootCmd.Flags().StringVar(&region, "region", "", "Region to capture: x,y,width,height")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output filename (default: screenshot_TIMESTAMP.png)")
	rootCmd.Flags().StringVarP(&display, "display", "d", "", "X11 display (default: $DISPLAY or :0)")
	rootCmd.Flags().BoolVarP(&listMon, "list", "l", false, "List available monitors")
	rootCmd.Flags().CountVarP(&compressLevel, "compress", "c", "Compression level: -c fast, -cc medium, -ccc best")
	rootCmd.Flags().BoolVarP(&raw, "raw", "r", false, "No compression (fastest, largest files)")
	rootCmd.Flags().BoolVarP(&view, "view", "v", false, "Open screenshot in default viewer after capture")
	rootCmd.Flags().BoolVar(&stdout, "stdout", false, "Output PNG to stdout (for piping)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	capturer := capture.New()

	// List monitors mode
	if listMon {
		return listMonitors(capturer)
	}

	// Determine output path
	outputPath := output
	if len(args) > 0 {
		outputPath = args[0]
	}
	if outputPath == "" {
		outputPath = capture.GenerateFilename("screenshot")
	}

	// Build capture options
	opts := strategy.CaptureOptions{
		Monitor: monitor,
		Display: display,
	}


	// Parse region if specified
	if region != "" {
		rect, err := parseRegion(region)
		if err != nil {
			return fmt.Errorf("invalid region: %w", err)
		}
		opts.Region = rect
	}

	// Determine compression level
	level := getCompressionLevel()

	// Stdout mode - output PNG directly to stdout
	if stdout {
		img, err := capturer.Capture(opts)
		if err != nil {
			return fmt.Errorf("capture failed: %w", err)
		}
		return capture.WritePNG(img, os.Stdout, level)
	}

	// Capture to file
	if err := capturer.CaptureToFile(opts, outputPath, level); err != nil {
		return err
	}

	fmt.Printf("Screenshot saved: %s\n", outputPath)

	// Open in viewer if requested
	if view {
		if err := openFile(outputPath); err != nil {
			return fmt.Errorf("failed to open viewer: %w", err)
		}
	}

	return nil
}

// openFile opens a file with the system's default application
func openFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Don't wait for the viewer to close
	return cmd.Start()
}

func listMonitors(capturer *capture.Capturer) error {
	monitors, err := capturer.ListMonitors()
	if err != nil {
		return err
	}

	fmt.Printf("Available monitors (%d):\n", len(monitors))
	for _, m := range monitors {
		fmt.Printf("  %d: %s (%dx%d at %d,%d)\n",
			m.Index,
			m.Name,
			m.Bounds.Dx(),
			m.Bounds.Dy(),
			m.Bounds.Min.X,
			m.Bounds.Min.Y,
		)
	}
	return nil
}

// parseRegion parses a region string "x,y,width,height" into an image.Rectangle
func parseRegion(s string) (*image.Rectangle, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 4 {
		return nil, fmt.Errorf("expected x,y,width,height")
	}

	vals := make([]int, 4)
	for i, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", p)
		}
		vals[i] = v
	}

	x, y, w, h := vals[0], vals[1], vals[2], vals[3]
	rect := image.Rect(x, y, x+w, y+h)
	return &rect, nil
}

// getCompressionLevel returns the compression level based on flags
// -r = NoCompression (0), -c = BestSpeed (1), -cc = DefaultCompression (2), -ccc = BestCompression (3)
func getCompressionLevel() int {
	if raw {
		return 0 // NoCompression
	}
	if compressLevel == 0 {
		return 1 // Default to BestSpeed (-c)
	}
	if compressLevel >= 3 {
		return 3 // BestCompression
	}
	return compressLevel // 1=BestSpeed, 2=DefaultCompression
}
