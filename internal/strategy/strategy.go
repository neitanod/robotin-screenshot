package strategy

import (
	"image"
)

// CaptureOptions holds the options for a screenshot capture
type CaptureOptions struct {
	// Monitor index (0-based). -1 means all monitors
	Monitor int

	// Region to capture. If nil, captures the full monitor/screen
	Region *image.Rectangle

	// WindowID to capture (X11 window ID). 0 means no specific window
	WindowID uint64

	// Display override (e.g., ":0"). Empty means use DISPLAY env var
	Display string
}

// Strategy defines the interface for screenshot capture strategies
type Strategy interface {
	// Name returns the strategy name (e.g., "x11", "wayland")
	Name() string

	// Available checks if this strategy can be used in the current environment
	Available() bool

	// Capture takes a screenshot with the given options
	Capture(opts CaptureOptions) (image.Image, error)

	// ListMonitors returns the available monitors
	ListMonitors() ([]Monitor, error)
}

// Monitor represents a display monitor
type Monitor struct {
	Index  int
	Name   string
	Bounds image.Rectangle
}
