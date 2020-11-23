package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"net/http"
	"strconv"
	"strings"

	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/gorilla/mux"
)

// For DICOM-specific needs

func (h *handler) TraceOverlayCINE(w http.ResponseWriter, r *http.Request) {
	// Fetch the desired images from the zip file and series. Either we can
	// reach into the primary manifest based on manifest_index, or if zip and
	// series are provided, we can use them directly.
	var zipFile, series string

	if mux.Vars(r)["manifest_index"] != "" {

		manifestIdx := mux.Vars(r)["manifest_index"]
		manifestIndex, err := strconv.Atoi(manifestIdx)
		if err != nil {
			HTTPError(h, w, r, fmt.Errorf("No manifest_index passed"))
			return
		}

		if manifestIndex >= len(h.Global.Manifest()) {
			HTTPError(h, w, r, fmt.Errorf("Manifest_index was %d, out of range of the Manifest slice", manifestIndex))
			return
		}

		manifestEntry := h.Global.Manifest()[manifestIndex]
		zipFile = manifestEntry.Zip
		series = manifestEntry.Series

		if cineBulkPath == "" || cineManifestPath == "" {
			HTTPError(h, w, r, fmt.Errorf("This program is not CINE-enabled"))
			return
		}
	} else {
		zipFile = mux.Vars(r)["zip"]
		series = mux.Vars(r)["series"]
	}

	r.ParseForm()

	showAll := false
	if r.FormValue("all") != "" {
		showAll = true
	}

	cineManifestPath = strings.TrimSuffix(cineManifestPath, "/")

	var dicomNames []string
	var seriesAlignment []string
	var err error
	if showAll {
		// Find all dicoms with the same Zip
		dicomNames, seriesAlignment, err = CINEFetchAllDicomNames(zipFile)
	} else {
		// Find all dicoms with the same Zip and Series
		dicomNames, err = CINEFetchDicomNames(zipFile, series)
	}
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Fetch actual images from the Zip
	client, err := getGSClient()
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}
	imageMap, err := bulkprocess.FetchNamedImagesFromZIP(cineBulkPath+"/"+zipFile, false, client, dicomNames)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Total hack to get the SAX images to display as a grid instead of
	// sequentially
	if showAll {
		newDicomNames, newImageMap, err := ImageGrid(dicomNames, imageMap, series, seriesAlignment, 50, 4)
		if err != nil {
			HTTPError(h, w, r, fmt.Errorf("%s:%s: %v", zipFile, series, err))
			return
		}

		// Override
		dicomNames = newDicomNames
		imageMap = newImageMap
	}

	// Create the GIF
	outGIF, err := bulkprocess.MakeOneGIFFromMap(dicomNames, imageMap, 2)
	if err != nil {
		HTTPError(h, w, r, fmt.Errorf("%s:%s: %v", zipFile, series, err))
		return
	}

	// Send the GIF over the wire
	if err := gif.EncodeAll(w, outGIF); err != nil {
		HTTPError(h, w, r, fmt.Errorf("%s:%s: %v", zipFile, series, err))
		return
	}
}

// ImageGrid is a hack that converts a SAX stack (structured as all 50
// timepoints for series 1, all 50 timepoints for series 2, etc) into a grid
// that simultaneously shows each series side-by-side. It highlights the active
// series based on a seriesAlignment slice that is paired 1:1 with the
// dicomNames slice.
func ImageGrid(dicomNames []string, imageMap map[string]image.Image, series string, seriesAlignment []string, imagesInCardiacCycle, Ncols int) ([]string, map[string]image.Image, error) {
	Nseries := len(imageMap) / imagesInCardiacCycle
	Nrows := Nseries / Ncols
	if Nseries%Ncols != 0 {
		Nrows++
	}
	maxWidth := -1
	maxHeight := -1
	for _, img := range imageMap {
		if x := img.Bounds().Dx(); x > maxWidth {
			maxWidth = x
		}
		if y := img.Bounds().Dy(); y > maxWidth {
			maxHeight = y
		}
	}

	newDicomNames := make([]string, 0, imagesInCardiacCycle)
	newImageMap := make(map[string]image.Image)

	// For each part of the cardiac cycle, draw our images, for a total of
	// imagesInCardiacCycle images
	for i := 0; i < imagesInCardiacCycle; i++ {

		r := image.Rect(0, 0, Ncols*maxWidth, Nrows*maxHeight)
		thisImg := image.NewRGBA(r)
		// Set a black background
		draw.Draw(thisImg, thisImg.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)

		seriesCounter := 0

		// We know that we need to draw Nseries images onto this template
	ImageGridLoop:
		for row := 0; row < Nrows; row++ {
			for col := 0; col < Ncols; col++ {

				imageID := i + imagesInCardiacCycle*(row*Ncols+col)
				if imageID > len(dicomNames) {
					continue
				}

				pane, exists := imageMap[dicomNames[imageID]]
				if !exists {
					break
				}

				startX := col * maxWidth
				startY := row * maxHeight

				drawRect := image.Rect(startX, startY, startX+pane.Bounds().Dx(), startY+pane.Bounds().Dy())

				// Draw the pane into the master image for this timepoint at its designated spot
				draw.Draw(thisImg, drawRect, pane, image.ZP, draw.Src)

				// Highlight the active pane
				if len(seriesAlignment) > imageID && seriesAlignment[imageID] == series {
					// Top
					innerRect := drawRect
					innerRect.Max.Y = innerRect.Min.Y + 2
					draw.Draw(thisImg, innerRect, &image.Uniform{color.White}, image.ZP, draw.Src)

					// Bottom
					innerRect = drawRect
					innerRect.Min.Y = innerRect.Max.Y - 2
					draw.Draw(thisImg, innerRect, &image.Uniform{color.White}, image.ZP, draw.Src)

					// Left
					innerRect = drawRect
					innerRect.Max.X = innerRect.Min.X + 2
					draw.Draw(thisImg, innerRect, &image.Uniform{color.White}, image.ZP, draw.Src)

					// Right
					innerRect = drawRect
					innerRect.Min.X = innerRect.Max.X - 2
					draw.Draw(thisImg, innerRect, &image.Uniform{color.White}, image.ZP, draw.Src)
				}

				seriesCounter++

				// Account for jagged series counts
				if seriesCounter >= Nseries {
					break ImageGridLoop
				}

			}
		}

		newDicomNames = append(newDicomNames, strconv.Itoa(i))
		newImageMap[strconv.Itoa(i)] = thisImg
	}

	return newDicomNames, newImageMap, nil
}
