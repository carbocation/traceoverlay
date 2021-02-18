package main

import (
	"fmt"
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

	gridColumns := 5
	if cols := r.FormValue("cols"); cols != "" {
		if intval, err := strconv.Atoi(cols); err == nil {
			gridColumns = intval
		}
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
		newDicomNames, newImageMap, err := bulkprocess.ImageGrid(dicomNames, imageMap, series, seriesAlignment, 50, gridColumns)
		if err != nil {
			HTTPError(h, w, r, fmt.Errorf("%s:%s: %v", zipFile, series, err))
			return
		}

		// Override
		dicomNames = newDicomNames
		imageMap = newImageMap
	}

	// Create the GIF
	outGIF, err := bulkprocess.MakeOneGIFFromMap(dicomNames, imageMap, 2, false)
	if err != nil {
		HTTPError(h, w, r, fmt.Errorf("%s:%s: %v", zipFile, series, err))
		return
	}

	// Set our content type
	w.Header().Set("Content-Type", "image/gif")

	// Send the GIF over the wire
	if err := gif.EncodeAll(w, outGIF); err != nil {
		HTTPError(h, w, r, fmt.Errorf("%s:%s: %v", zipFile, series, err))
		return
	}
}
