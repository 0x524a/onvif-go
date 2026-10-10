package server

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"time"
)

const (
	fallbackSnapshotWidth  = 640
	fallbackSnapshotHeight = 360
	maxSnapshotDimension   = 4096
	defaultJPEGQuality     = 80
	secondsPerMinute       = 60
	barsFraction           = 3 // color bars fill the top 2/3 of the frame
	markerHeightDivisor    = 12

	// Studio-range values for the color bars, and full-scale for the ramp.
	barHigh   = 235
	barLow    = 16
	fullScale = 255
)

// testBars are the seven bars of a standard color-bar test pattern.
var testBars = [...]color.RGBA{
	{R: barHigh, G: barHigh, B: barHigh, A: fullScale}, // white
	{R: barHigh, G: barHigh, B: barLow, A: fullScale},  // yellow
	{R: barLow, G: barHigh, B: barHigh, A: fullScale},  // cyan
	{R: barLow, G: barHigh, B: barLow, A: fullScale},   // green
	{R: barHigh, G: barLow, B: barHigh, A: fullScale},  // magenta
	{R: barHigh, G: barLow, B: barLow, A: fullScale},   // red
	{R: barLow, G: barLow, B: barHigh, A: fullScale},   // blue
}

// renderSnapshot returns a JPEG test pattern sized for the profile's snapshot
// configuration. It stands in for a camera frame: a color-bar chart above a
// grey ramp, with a white marker whose position follows the second of the
// minute, so two snapshots taken a moment apart can be told apart.
func renderSnapshot(profile *ProfileConfig, now time.Time) ([]byte, error) {
	width, height := snapshotSize(profile)

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	barsEnd := height * (barsFraction - 1) / barsFraction

	for y := 0; y < height; y++ {
		row := img.Pix[y*img.Stride : y*img.Stride+width*4]

		for x := 0; x < width; x++ {
			var c color.RGBA
			if y < barsEnd {
				c = testBars[x*len(testBars)/width]
			} else {
				g := uint8(x * fullScale / max(width-1, 1)) // #nosec G115 -- x <= width-1, so the value is at most 255
				c = color.RGBA{R: g, G: g, B: g, A: fullScale}
			}

			row[x*4], row[x*4+1], row[x*4+2], row[x*4+3] = c.R, c.G, c.B, c.A
		}
	}

	drawMarker(img, now)

	quality := int(profile.Snapshot.Quality)
	if quality <= 0 || quality > 100 {
		quality = defaultJPEGQuality
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode snapshot: %w", err)
	}

	return buf.Bytes(), nil
}

// snapshotSize picks the frame size: the snapshot resolution, then the video
// source resolution, then a small default. Absurd sizes are clamped so a bad
// config cannot make a request allocate gigabytes.
func snapshotSize(profile *ProfileConfig) (width, height int) {
	width, height = profile.Snapshot.Resolution.Width, profile.Snapshot.Resolution.Height
	if width <= 0 || height <= 0 {
		width, height = profile.VideoSource.Resolution.Width, profile.VideoSource.Resolution.Height
	}

	if width <= 0 || height <= 0 {
		return fallbackSnapshotWidth, fallbackSnapshotHeight
	}

	return min(width, maxSnapshotDimension), min(height, maxSnapshotDimension)
}

// drawMarker paints a white block along the bottom whose x position is the
// current second of the minute.
func drawMarker(img *image.RGBA, now time.Time) {
	b := img.Bounds()
	size := max(b.Dy()/markerHeightDivisor, 1)
	x0 := now.Second() * (b.Dx() - size) / (secondsPerMinute - 1)

	white := color.RGBA{R: fullScale, G: fullScale, B: fullScale, A: fullScale}
	for y := b.Dy() - size; y < b.Dy(); y++ {
		for x := x0; x < x0+size && x < b.Dx(); x++ {
			img.SetRGBA(x, y, white)
		}
	}
}
