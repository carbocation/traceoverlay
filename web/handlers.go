package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/carbocation/genomisc/overlay"
	"github.com/carbocation/genomisc/ukbb/bulkprocess"
	"github.com/gorilla/mux"
	"golang.org/x/image/bmp"
)

func (h *handler) TemplateOnly(w http.ResponseWriter, r *http.Request) {
	tpl := mux.Vars(r)["template"]
	if tpl == "" {
		tpl = "index"
	}

	Render(h, w, r, strings.Title(tpl), fmt.Sprintf("%s.html", tpl), nil, nil)
}

func (h *handler) Index(w http.ResponseWriter, r *http.Request) {
	output := struct{ Project string }{h.Global.Project}

	Render(h, w, r, h.Global.Site, "index.html", output, nil)
}

func (h *handler) ListProject(w http.ResponseWriter, r *http.Request) {
	if err := UpdateManifest(); err != nil {
		HTTPError(h, w, r, err)
		return
	}

	output := struct {
		Project  string
		Manifest []Manifest
	}{
		h.Global.Project,
		h.Global.Manifest(),
	}

	Render(h, w, r, "List Project", "listproject.html", output, nil)
}

func (h *handler) TraceOverlay(w http.ResponseWriter, r *http.Request) {
	// Fetch the desired image from the zip file as described in the manifest
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
	var im image.Image

	if strings.HasPrefix(h.Global.Config.ImagePath, "gs://") {
		// Google Storage
		client, err := getGSClient()
		if err != nil {
			HTTPError(h, w, r, err)
			return
		}
		im, err = bulkprocess.ExtractDicomFromGoogleStorage(fmt.Sprintf("%s/%s", h.Global.Config.ImagePath, manifestEntry.Zip),
			manifestEntry.Dicom,
			true,
			client)
	} else if h.Global.Config.PreParsed && !strings.HasPrefix(h.Global.Config.ImagePath, "gs://") {
		im, err = bulkprocess.ExtractImageFromLocalFile(manifestEntry.Dicom, h.Global.Config.ImageSuffix, h.Global.Config.ImagePath)
	} else {
		pathPart := path.Dir(h.Global.ManifestPath)
		im, err = bulkprocess.ExtractDicomFromLocalFile(fmt.Sprintf("%s/%s", pathPart, manifestEntry.Zip), manifestEntry.Dicom, true)
	}
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	var encodedOverlayString string
	if manifestEntry.HasOverlayFromProject {
		switch {
		default:
			pngPath := filepath.Join(global.Config.LabelPath, manifestEntry.OverlayFilename())
			f, err := os.Open(pngPath)
			if err != nil {
				break
			}

			img, err := png.Decode(f)
			if err != nil {
				break
			}

			// Decode into something visually acceptable
			humanImg, err := h.Config.Labels.DecodeImageFromImageSegment(img, true)
			if err != nil {
				HTTPError(h, w, r, err)
				return
			}

			var imBuff bytes.Buffer
			png.Encode(&imBuff, humanImg)
			// png.Encode(&imBuff, img)
			encodedOverlayString = base64.StdEncoding.EncodeToString(imBuff.Bytes())
		}
	}

	// Convert that image to a PNG and base64 encode it so we can show it raw
	var imBuff bytes.Buffer
	png.Encode(&imBuff, im)
	encodedString := base64.StdEncoding.EncodeToString(imBuff.Bytes())

	output := struct {
		Project             string
		ManifestEntry       Manifest
		ManifestIndex       int
		EncodedImage        string
		Width               int
		Height              int
		HasOverlay          bool
		EncodedOverlayImage string
		DefaultBrush        string
		BrushSize           int
		Labels              []overlay.Label
	}{
		h.Global.Project,
		manifestEntry,
		manifestIndex,
		strings.NewReplacer("\n", "", "\r", "").Replace(encodedString),
		im.Bounds().Dx(),
		im.Bounds().Dy(),
		manifestEntry.HasOverlayFromProject,
		strings.NewReplacer("\n", "", "\r", "").Replace(encodedOverlayString),
		h.Config.DefaultBrush,
		h.Config.BrushSize,
		h.Config.Labels.Sorted(),
	}

	Render(h, w, r, "Trace Overlay", "traceoverlay.html", output, nil)
}

func (h *handler) TraceOverlayPost(w http.ResponseWriter, r *http.Request) {
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

	nextManifestIndex := manifestIndex + 1
	if len(h.Global.Manifest()) <= nextManifestIndex {
		// Back to the start if you roll over
		nextManifestIndex = 1
	}

	output := struct {
		Project           string
		ManifestIndex     int
		ManifestEntry     Manifest
		EncodedImage      string
		Width             int
		Height            int
		NextManifestIndex int
	}{
		Project:           h.Global.Project,
		ManifestIndex:     manifestIndex,
		ManifestEntry:     manifestEntry,
		NextManifestIndex: nextManifestIndex,
	}

	// Permit uploads up to 128 megabytes (go default is 10 megabytes)
	r.Body = http.MaxBytesReader(w, r.Body, 128*1024*1024)

	parsedImage := strings.SplitAfterN(r.PostFormValue("imgBase64"), ",", 2)
	if len(parsedImage) < 2 {
		log.Println(r.FormValue("imgBase64"))
		HTTPError(h, w, r, fmt.Errorf("Parsed image had fewer parts than expected"))
		return
	}

	imgRdr := strings.NewReader(parsedImage[1])
	dec := base64.NewDecoder(base64.StdEncoding, imgRdr)

	bmpImage, err := bmp.Decode(dec)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Generate an encoded version of the image that uses #000000 for
	// background, #010101 for ID 1, etc
	encoded, err := h.Config.Labels.EncodeImageToImageSegment(bmpImage)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Save the BMP to disk under your project folder, using the ID-encoding.
	f, err := os.Create(filepath.Join(global.Config.LabelPath, manifestEntry.OverlayFilename()))
	if err != nil {
		HTTPError(h, w, r, err)
	}
	defer f.Close()

	// Write the PNG representation of our ID-encoded image to disk
	if err := png.Encode(f, encoded); err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Decode the ID-encoding we just created into a human-interpretable set of
	// colors.
	decoded, err := h.Config.Labels.DecodeImageFromImageSegment(encoded, true)
	if err != nil {
		HTTPError(h, w, r, err)
		return
	}

	// Convert our (decoded, human-readable) image mask to a PNG and base64
	// encode it so we can show it raw
	var imBuff bytes.Buffer

	// For now, instead of the submitted image, we will display (on the reponse
	// page) the decoding of the RLE-encoded version of the mask as proof that
	// we can encode/decode properly.
	// png.Encode(&imBuff, bmpImage)
	png.Encode(&imBuff, decoded)

	base64Image := base64.StdEncoding.EncodeToString(imBuff.Bytes())

	output.EncodedImage = base64Image
	output.Width = decoded.Bounds().Dx()
	output.Height = decoded.Bounds().Dy()

	Render(h, w, r, "Trace Overlay", "traceoverlay-POST.html", output, nil)
}

func (h *handler) Goroutines(w http.ResponseWriter, r *http.Request) {
	goroutines := fmt.Sprintf("%d goroutines are currently active\n", runtime.NumGoroutine())

	w.Write([]byte(goroutines))
}
