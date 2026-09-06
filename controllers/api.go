package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/z-wentao/PhotoShare/context"
	"github.com/z-wentao/PhotoShare/models"
)

// A hard ceiling on request size. ParseMultipartForm's argument is only a
// memory threshold — anything larger spills to a temp file — so without this
// there is no upper bound at all.
const maxUploadBytes = 20 << 20 // 20MB

type API struct {
	APITokenService *models.APITokenService
	ImageService    *models.ImageService
	BaseURL         string
}

type uploadResponse struct {
	URL      string `json:"url"`
	Markdown string `json:"markdown"`
	Filename string `json:"filename"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		fmt.Println(err)
	}
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// RequireToken authenticates from the Authorization header and nothing else.
// It must never fall back to the session cookie: browsers do not attach this
// header on their own, and that is the sole reason this endpoint can sit
// outside CSRF protection.
func (a API) RequireToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, ok := bearerToken(r)
		if !ok {
			writeJSONError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		token, err := a.APITokenService.ByToken(raw)
		if err != nil {
			if !errors.Is(err, models.ErrNotFound) {
				fmt.Println(err)
			}
			writeJSONError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := context.WithAPIToken(r.Context(), token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

// Upload takes no gallery parameter: the destination comes from the token, so
// there is no gallery ownership check to write, because there is nothing the
// client could have forged.
func (a API) Upload(w http.ResponseWriter, r *http.Request) {
	token := context.APIToken(r.Context())
	if token == nil {
		writeJSONError(w, http.StatusUnauthorized, "missing bearer token")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	err := r.ParseMultipartForm(5 << 20)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeJSONError(w, http.StatusRequestEntityTooLarge, "file too large")
			return
		}
		writeJSONError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, `missing "file" field`)
		return
	}
	defer file.Close()

	img, err := a.ImageService.Create(token.GalleryID, header.Filename, file)
	if err != nil {
		var fileErr models.FileError
		if errors.As(err, &fileErr) {
			writeJSONError(w, http.StatusBadRequest, fileErr.Error())
			return
		}
		fmt.Println(err)
		writeJSONError(w, http.StatusInternalServerError, "upload failed")
		return
	}

	url := fmt.Sprintf("%s/i/%s%s", strings.TrimRight(a.BaseURL, "/"), img.UUID, img.Ext)
	writeJSON(w, http.StatusOK, uploadResponse{
		URL:      url,
		Markdown: fmt.Sprintf("![%s](%s)", img.OrigName, url),
		Filename: img.OrigName,
	})
}
