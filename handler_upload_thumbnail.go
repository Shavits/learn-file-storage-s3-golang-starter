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

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusBadRequest, "Unable to parse file", err)
	// 	return
	// }

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Unable to find video", err)
		return
	}

	if video.CreateVideoParams.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Video does not belong to current user", err)
		return
	}

	exts, _ := mime.ExtensionsByType(mediaType)
	ext := ""
	if len(exts) > 0 {
		ext = exts[0]
	}
	fileName := fmt.Sprintf("%v%v", videoID, ext)

	path := filepath.Join(cfg.assetsRoot, fileName)
	fmt.Print(path)
	created, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to create thumbnail file", err)
		return
	}
	defer created.Close()

	_, err = io.Copy(created, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to save thumbnail file", err)
		return
	}
	url := fmt.Sprintf("http://localhost:%v/%v", cfg.port, path)
	video.ThumbnailURL = &url
	fmt.Print(video.ThumbnailURL)
	cfg.db.UpdateVideo(video)

	respondWithJSON(w, http.StatusOK, video)
}
