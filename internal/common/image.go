package common

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"time"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// RenderImage fetches an image from a URL and returns a terminal-renderable ANSI halfblock string.
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

	src, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("decode image: %w", err)
	}

	bounds := src.Bounds()
	imgW, imgH := float64(bounds.Dx()), float64(bounds.Dy())
	if imgW == 0 || imgH == 0 {
		return "", fmt.Errorf("invalid image dimensions")
	}

	ar := imgW / imgH // image aspect ratio
	w := float64(maxWidth)
	h := w / (ar * 2.0) // convert to cell rows (halfblocks -> 2 vertical pixels per cell)
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

	pixelW := cellW
	pixelH := cellH * 2
	dst := image.NewRGBA(image.Rect(0, 0, pixelW, pixelH))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	var sb strings.Builder
	for y := 0; y < pixelH; y += 2 {
		for x := 0; x < pixelW; x++ {
			r1, g1, b1, a1 := dst.At(x, y).RGBA()
			r2, g2, b2, a2 := dst.At(x, y+1).RGBA()

			// RGBA returns 0..65535, scale to 0..255
			r1, g1, b1, a1 = r1>>8, g1>>8, b1>>8, a1>>8
			r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8

			if a1 < 128 && a2 < 128 {
				sb.WriteString(" ")
			} else if a1 < 128 {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▄\x1b[0m", r2, g2, b2))
			} else if a2 < 128 {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▀\x1b[0m", r1, g1, b1))
			} else {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m", r1, g1, b1, r2, g2, b2))
			}
		}
		if y+2 < pixelH {
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}
