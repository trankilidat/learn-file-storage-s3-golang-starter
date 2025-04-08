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

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	fmt.Println("parsed multi part form")

	// "thumbnail" should match the HTML form input name
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	extensions, err := mime.ExtensionsByType(mediaType)
	if err != nil || len(extensions) == 0 {
		respondWithError(w, http.StatusBadRequest, "Could not determine file extension from media type", err)
		return
	}
	ext := extensions[0] // e.g. ".png"

	videoPath := fmt.Sprintf(`%s.%s`, videoID, ext)
	path := filepath.Join(cfg.assetsRoot, videoPath)

	fl, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to create file", err)
		return
	}

	_, err = io.Copy(fl, file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to copy file to file system", err)
		return
	}

	// get video metadata
	vm, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to get video", err)
		return
	}

	if vm.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User is not authorized to access video", err)
		return
	}

	tnu := fmt.Sprintf(`http://localhost:8091/assets/%s`, videoPath)
	vm.ThumbnailURL = &tnu

	// update video with new thumbnail url
	if err := cfg.db.UpdateVideo(vm); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vm)
}
