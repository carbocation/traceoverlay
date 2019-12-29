package main

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strconv"
	"strings"

	"github.com/tj/go-rle"
)

type Label struct {
	Label string
	ID    uint   `json:"id"`
	Color string `json:"color"`
}

type LabelMap map[string]Label

func (l LabelMap) EncodeImageToRLE(bmpImage image.Image) ([]byte, error) {
	// Map from the color code to the Label, with all of its attached
	// information
	colorLabels := make(map[color.Color]Label)
	pixelLabels := make([]int64, 0)

	for y := 0; y < bmpImage.Bounds().Max.Y; y++ {
		for x := 0; x < bmpImage.Bounds().Max.X; x++ {

			// Find the color at this pixel
			c := bmpImage.At(x, y)

			// If we haven't yet mapped this point's color to a Label
			// identifier, do so now:
			if _, exists := colorLabels[c]; !exists {

				// Create the hex string representation
				r, g, b, _ := c.RGBA()
				cl := fmt.Sprintf("#%02x%02x%02x", uint8(r), uint8(g), uint8(b))

				// Look up the hex string representation and map it
				for _, v := range l.Sorted() {
					if v.Color == cl {
						colorLabels[c] = v
						break
					} else if cl == "#000000" && v.ID == 0 {
						// Background, special case
						colorLabels[c] = v
						break
					}
				}
			}

			// Make sure that all labels are known
			lab, exists := colorLabels[c]
			if !exists {
				return nil, fmt.Errorf("Saw color %v but could not find this color in the label map", c)
			}
			pixelLabels = append(pixelLabels, int64(lab.ID))
		}
	}

	encoded := rle.EncodeInt64(pixelLabels)

	return encoded, nil
}

func (l LabelMap) DecodeImageFromRLE(rleBytes []byte, maxX, maxY int) (image.Image, error) {
	slc, err := rle.DecodeInt64(rleBytes)
	if err != nil {
		return nil, err
	}

	// Know which label maps to each integer
	labelColors := make(map[int64]Label)
	for _, v := range l {
		labelColors[int64(v.ID)] = v
	}

	// Know which color maps to each entry
	colorCodes := make([]string, 0, len(slc))
	for _, v := range slc {
		colorCodes = append(colorCodes, labelColors[v].Color)
	}

	// Paint each pixel
	img := image.NewRGBA(image.Rect(0, 0, maxX, maxY))

	for i, label := range slc {
		colorCode := labelColors[label].Color

		colHere, err := rgbaFromColorCode(colorCode)
		if err != nil {
			return nil, err
		}

		// img.Set(i/maxY, i%maxY, colHere)
		img.Set(i%maxX, i/maxX, colHere)
	}

	return img, nil
}

// Valid ensures that the LabelMap is valid by testing that it is bijective,
// starts with 0, and has no gaps. If not, it's invalid.
func (l LabelMap) Valid() bool {
	inverse := make(map[uint]string)
	for k, v := range l {
		inverse[v.ID] = k
	}

	// Bijective?
	if !(len(l) == len(inverse)) {
		return false
	}

	// Starts with 0 and has consecutive integers?
	for i := 0; i < len(inverse); i++ {
		if _, exists := inverse[uint(i)]; !exists {
			return false
		}
	}

	return true
}

func (l LabelMap) Sorted() []Label {
	out := make([]Label, 0, len(l))

	for k, v := range l {
		v.Label = k
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[j].ID > out[i].ID
	})

	return out
}

func rgbaFromColorCode(colorCode string) (color.Color, error) {
	colorCode = strings.ReplaceAll(colorCode, "#", "")

	// Special case the background
	if len(colorCode) < 6 {
		return color.RGBA{0, 0, 0, 0}, nil
	}

	// Parse each channel
	r, err := strconv.ParseInt(colorCode[0:2], 16, 16)
	if err != nil {
		return nil, err
	}
	g, err := strconv.ParseInt(colorCode[2:4], 16, 16)
	if err != nil {
		return nil, err
	}
	b, err := strconv.ParseInt(colorCode[4:6], 16, 16)
	if err != nil {
		return nil, err
	}

	return color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}, nil
}
