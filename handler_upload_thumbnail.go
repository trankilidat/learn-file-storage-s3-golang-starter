package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

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

	fmt.Println("mediatype", mediaType)

	// read image into byte slice
	imgSlice, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to read image data into byte slice", err)
		return
	}

	// base64 encode image so we can store it as text in db
	encodedImage := base64.StdEncoding.EncodeToString(imgSlice)
	dataURL := fmt.Sprintf(`data:%s;base64,%s`, mediaType, encodedImage)

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

	// add dataurl in thumnail url to video metadata
	vm.ThumbnailURL = &dataURL

	// update video with new thumbnail url
	if err := cfg.db.UpdateVideo(vm); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, vm)
}
