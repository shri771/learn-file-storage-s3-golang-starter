package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	const maxMemory = 10 << 20 // 10 MB
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	// Try to get content type from header first; if empty or ambiguous, sniff it from file bytes.
	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		respondWithError(w, http.StatusInternalServerError, "Uploaded file is not seekable", nil)
		return
	}

	exts, _ := mime.ExtensionsByType(mediaType)
	ext := ""
	if len(exts) > 0 {
		ext = exts[0] // e.g. ".png"
	} else {
		// fallback extension if unknown
		ext = ".bin"
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	// Check for correct file format
	if ext == ".pdf" {
		respondWithError(w, http.StatusBadRequest, ".pdf not supported in thumbnails", err)
		return
	}

	// Build the path and ensure directory exists
	filePath := cfg.GetFilePath(video.UserID, ext) // ensure this returns a full path
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create directories for thumbnail", err)
		return
	}

	dst, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create thumbnail file", err)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not save thumbnail", err)
		return
	}

	// Build URL (use filepath.Base so assets mapping is simple). Prefer to compute public path separately.
	filename := filepath.Base(filePath)
	publicPath := filepath.ToSlash(filepath.Join("assets", filename)) // path for URL
	thumbnailURL := fmt.Sprintf("http://localhost:%v/%s", cfg.port, publicPath)
	video.ThumbnailURL = &thumbnailURL

	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
