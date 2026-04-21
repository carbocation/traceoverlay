package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var inferenceClient = &http.Client{Timeout: 120 * time.Second}

// PredictModels proxies GET /models from the inference service.
func (h *handler) PredictModels(w http.ResponseWriter, r *http.Request) {
	if h.InferenceURL == "" {
		http.Error(w, "inference not configured", http.StatusNotFound)
		return
	}
	resp, err := inferenceClient.Get(h.InferenceURL + "/models")
	if err != nil {
		http.Error(w, fmt.Sprintf("inference service unavailable: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// Predict fetches the image for the given manifest index, sends it to the
// inference service, and returns the raw-category mask PNG (base64) plus the
// model's output_map so the client can perform the pixel merge itself.
func (h *handler) Predict(w http.ResponseWriter, r *http.Request) {
	if h.InferenceURL == "" {
		http.Error(w, "inference not configured", http.StatusNotFound)
		return
	}

	manifestIdx := mux.Vars(r)["manifest_index"]
	manifestIndex, err := strconv.Atoi(manifestIdx)
	if err != nil {
		http.Error(w, "invalid manifest_index", http.StatusBadRequest)
		return
	}
	if manifestIndex >= len(h.Global.Manifest()) {
		http.Error(w, "manifest_index out of range", http.StatusBadRequest)
		return
	}

	modelName := r.URL.Query().Get("model")
	if modelName == "" {
		http.Error(w, "model query parameter required", http.StatusBadRequest)
		return
	}

	im, err := h.fetchImage(manifestIndex)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, im); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build multipart/form-data body for the inference service.
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("image", "image.png")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(fw, &imgBuf); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := mw.WriteField("model", modelName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mw.Close()

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, h.InferenceURL+"/infer", &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, err := inferenceClient.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("inference service unavailable: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
