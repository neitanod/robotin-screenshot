//go:build linux

package strategy

import (
	"fmt"
	"image"
	"os"

	"github.com/kbinani/screenshot"
)

// X11Strategy implements screenshot capture for X11
type X11Strategy struct {
	originalDisplay string
}

// NewX11Strategy creates a new X11 screenshot strategy
func NewX11Strategy() *X11Strategy {
	return &X11Strategy{}
}

// Name returns the strategy name
func (s *X11Strategy) Name() string {
	return "x11"
}

// Available checks if X11 is available
func (s *X11Strategy) Available() bool {
	display := os.Getenv("DISPLAY")
	if display == "" {
		// Try to set a default display
		os.Setenv("DISPLAY", ":0")
		display = ":0"
	}

	// Check if we can get display count (quick availability check)
	n := screenshot.NumActiveDisplays()
	return n > 0
}

// setDisplay temporarily sets DISPLAY env var and returns a cleanup function
func (s *X11Strategy) setDisplay(display string) func() {
	if display == "" {
		return func() {}
	}

	s.originalDisplay = os.Getenv("DISPLAY")
	os.Setenv("DISPLAY", display)

	return func() {
		if s.originalDisplay != "" {
			os.Setenv("DISPLAY", s.originalDisplay)
		} else {
			os.Unsetenv("DISPLAY")
		}
	}
}

// ensureDisplay makes sure DISPLAY is set, using fallback if needed
func (s *X11Strategy) ensureDisplay(opts CaptureOptions) func() {
	// If explicit display requested, use it
	if opts.Display != "" {
		return s.setDisplay(opts.Display)
	}

	// If DISPLAY not set, use default :0 (for cron/lockscreen)
	if os.Getenv("DISPLAY") == "" {
		return s.setDisplay(":0")
	}

	return func() {}
}

// Capture takes a screenshot
func (s *X11Strategy) Capture(opts CaptureOptions) (image.Image, error) {
	cleanup := s.ensureDisplay(opts)
	defer cleanup()

	// If a specific region is requested
	if opts.Region != nil {
		return screenshot.CaptureRect(*opts.Region)
	}

	// Get number of displays
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, fmt.Errorf("no active displays found")
	}

	// Capture all monitors combined
	if opts.Monitor == -1 {
		// Calculate combined bounds
		var minX, minY, maxX, maxY int
		for i := 0; i < n; i++ {
			bounds := screenshot.GetDisplayBounds(i)
			if i == 0 || bounds.Min.X < minX {
				minX = bounds.Min.X
			}
			if i == 0 || bounds.Min.Y < minY {
				minY = bounds.Min.Y
			}
			if i == 0 || bounds.Max.X > maxX {
				maxX = bounds.Max.X
			}
			if i == 0 || bounds.Max.Y > maxY {
				maxY = bounds.Max.Y
			}
		}
		allBounds := image.Rect(minX, minY, maxX, maxY)
		return screenshot.CaptureRect(allBounds)
	}

	// Capture specific monitor
	if opts.Monitor < 0 || opts.Monitor >= n {
		return nil, fmt.Errorf("monitor %d out of range (0-%d)", opts.Monitor, n-1)
	}

	bounds := screenshot.GetDisplayBounds(opts.Monitor)
	return screenshot.CaptureRect(bounds)
}

// ListMonitors returns the available monitors
func (s *X11Strategy) ListMonitors() ([]Monitor, error) {
	// Ensure display is set
	if os.Getenv("DISPLAY") == "" {
		os.Setenv("DISPLAY", ":0")
	}

	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, fmt.Errorf("no active displays found")
	}

	monitors := make([]Monitor, n)
	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		monitors[i] = Monitor{
			Index:  i,
			Name:   fmt.Sprintf("Display %d", i),
			Bounds: bounds,
		}
	}

	return monitors, nil
}
