package main

import (
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"net/http"
	"strconv"
	"time"
)

const maxUploadSize = 30 << 20

func processHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large or malformed form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing 'image' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, "could not decode image", http.StatusBadRequest)
		return
	}

	filterName := r.PathValue("name")
	filterFn, ok := filterRegistry[filterName]
	if !ok {
		http.Error(w, "unknown filter: "+filterName, http.StatusBadRequest)
		return
	}
	result, _ := filterFn(img)

	duration := time.Since(start)
	w.Header().Set("X-Process-Time-Ms", strconv.FormatInt(duration.Milliseconds(), 10))
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, result); err != nil {
		http.Error(w, "failed to encode result", http.StatusInternalServerError)
		return
	}

	log.Printf("filter=%s duration=%s", filterName, duration)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
