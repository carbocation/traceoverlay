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

func manifestMap(h *handler) map[string]struct{} {
	// Ensure that we only
	annotationManifest := h.Manifest()
	annotationManifestMap := make(map[string]struct{})
	for _, v := range annotationManifest {
		annotationManifestMap[v.Zip] = struct{}{}
	}

	return annotationManifestMap
}

// TraceOverlayCINEHTML provides an HTML wrapper to fetch the TraceOverlayCINE.
// This can be nice because it allows you to use CSS to change the display of
// the image. In this case, we make the image fill the height or width of the
// viewport (whichever is smaller).
func (h *handler) TraceOverlayCINEHTML(w http.ResponseWriter, r *http.Request) {

	// TODO: proper URL query string construction
	output := struct {
		ImageLink string
	}{
		fmt.Sprintf("/traceoverlay/cine/%s/%s?all=%s", mux.Vars(r)["zip"], mux.Vars(r)["series"], r.FormValue("all")),
	}

	Render(h, w, r, "CINE", "cine.html", output, nil)
}

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

	manMap := manifestMap(h)

	if showAll {
		// Find all dicoms with the same Zip
		dicomNames, seriesAlignment, err = CINEFetchAllDicomNames(manMap, zipFile)
	} else {
		// Find all dicoms with the same Zip and Series
		dicomNames, err = CINEFetchDicomNames(manMap, zipFile, series)
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
