package common

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	"github.com/blacktop/go-termimg"
)

// RenderImage fetches an image from a URL and returns a terminal-renderable string.
// maxWidth and maxHeight are the maximum bounds in character cells.
// The image is scaled to fit within those bounds while preserving aspect ratio.
func RenderImage(url string, maxWidth, maxHeight int) (string, error) {
	if url == "" {
		return "", fmt.Errorf("empty URL")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetch image: status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()
	imgW, imgH := float64(bounds.Dx()), float64(bounds.Dy())
	if imgW == 0 || imgH == 0 {
		return "", fmt.Errorf("invalid image dimensions")
	}

	ar := imgW / imgH // image aspect ratio
	w := float64(maxWidth)
	h := w / (ar * 2.0) // convert to cell rows (halfblocks → *2)
	if h > float64(maxHeight) {
		h = float64(maxHeight)
		w = ar * h * 2.0
	}
	cellW := int(w)
	cellH := int(h)
	if cellW < 1 {
		cellW = 1
	}
	if cellH < 1 {
		cellH = 1
	}

	ti := termimg.New(img)
	ti.Width(cellW).Height(cellH).Scale(termimg.ScaleFit).Protocol(termimg.Halfblocks)

	rendered, err := ti.Render()
	if err != nil {
		return "", fmt.Errorf("render image: %w", err)
	}
	return rendered, nil
}
