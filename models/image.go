package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/z-wentao/PhotoShare/rand"
)

// 16 bytes = 128 bits of entropy, 22 chars once encoded. Shorter than a
// session token because leaking an image URL is far less severe than
// leaking a session.
const imageUUIDBytes = 16

type Image struct {
	ID        int
	UUID      string
	GalleryID int
	Ext       string // includes the dot, e.g. ".jpg"
	OrigName  string
}

type ImageService struct {
	DB        *sql.DB
	ImagesDir string
}

func (s *ImageService) Create(galleryID int, filename string, contents io.ReadSeeker) (*Image, error) {
	err := checkContentType(contents, s.imageContentTypes())
	if err != nil {
		return nil, fmt.Errorf("create image %v: %w", filename, err)
	}
	err = checkExtension(filename, s.extensions())
	if err != nil {
		return nil, fmt.Errorf("create image %v: %w", filename, err)
	}

	uuid, err := rand.URLSafeString(imageUUIDBytes)
	if err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}

	img := Image{
		UUID:      uuid,
		GalleryID: galleryID,
		Ext:       strings.ToLower(filepath.Ext(filename)),
		OrigName:  filename,
	}
	path := s.Path(img)

	// Write the file first, the DB record second. A crash in between leaves
	// an orphan file nobody can see, which is recoverable. The other order
	// would leave a row pointing at a missing file - a broken image the user
	// can see and cannot fix.
	err = os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return nil, fmt.Errorf("create image directory: %w", err)
	}
	dst, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create image file: %w", err)
	}
	_, err = io.Copy(dst, contents)
	if err != nil {
		dst.Close()
		os.Remove(path)
		return nil, fmt.Errorf("copy contents to image: %w", err)
	}
	// Close explicitly rather than deferring: it flushes buffered writes and
	// can fail, and we must not report success on a half-written file.
	err = dst.Close()
	if err != nil {
		os.Remove(path)
		return nil, fmt.Errorf("close image file: %w", err)
	}

	row := s.DB.QueryRow(`
		INSERT INTO images (uuid, gallery_id, ext, orig_name)
		VALUES ($1, $2, $3, $4) RETURNING id;`,
		img.UUID, img.GalleryID, img.Ext, img.OrigName)
	err = row.Scan(&img.ID)
	if err != nil {
		os.Remove(path)
		return nil, fmt.Errorf("insert image: %w", err)
	}
	return &img, nil
}

func (s *ImageService) ByGalleryID(galleryID int) ([]Image, error) {
	// id is monotonically increasing, so ordering by it descending gives
	// newest-first without needing a created_at column.
	rows, err := s.DB.Query(`
		SELECT id, uuid, gallery_id, ext, orig_name
		FROM images
		WHERE gallery_id = $1
		ORDER BY id DESC;`, galleryID)
	if err != nil {
		return nil, fmt.Errorf("images by gallery: %w", err)
	}
	defer rows.Close()

	var images []Image
	for rows.Next() {
		var img Image
		err := rows.Scan(&img.ID, &img.UUID, &img.GalleryID, &img.Ext, &img.OrigName)
		if err != nil {
			return nil, fmt.Errorf("images by gallery: %w", err)
		}
		images = append(images, img)
	}
	// rows.Next() returns false both when the data is exhausted and when the
	// query failed partway. Without this check a broken query looks like a
	// short but successful read.
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("images by gallery: %w", err)
	}
	return images, nil
}

func (s *ImageService) ByUUID(uuid string) (*Image, error) {
	var img Image
	row := s.DB.QueryRow(`
		SELECT id, uuid, gallery_id, ext, orig_name
		FROM images
		WHERE uuid = $1;`, uuid)
	err := row.Scan(&img.ID, &img.UUID, &img.GalleryID, &img.Ext, &img.OrigName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("image by uuid: %w", err)
	}
	return &img, nil
}

// Delete removes the DB record first, then the file. A crash in between
// leaves an orphan file rather than a row pointing at nothing.
func (s *ImageService) Delete(img Image) error {
	_, err := s.DB.Exec(`DELETE FROM images WHERE id = $1;`, img.ID)
	if err != nil {
		return fmt.Errorf("delete image: %w", err)
	}
	err = os.Remove(s.Path(img))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete image file: %w", err)
	}
	return nil
}

// DeleteByGalleryID removes every image belonging to a gallery, records and
// files both. The gallery itself is left alone.
func (s *ImageService) DeleteByGalleryID(galleryID int) error {
	_, err := s.DB.Exec(`DELETE FROM images WHERE gallery_id = $1;`, galleryID)
	if err != nil {
		return fmt.Errorf("delete images by gallery: %w", err)
	}
	err = os.RemoveAll(s.galleryDir(galleryID))
	if err != nil {
		return fmt.Errorf("delete gallery images dir: %w", err)
	}
	return nil
}

func (s *ImageService) imagesDir() string {
	if s.ImagesDir == "" {
		return "images"
	}
	return s.ImagesDir
}

func (s *ImageService) galleryDir(galleryID int) string {
	return filepath.Join(s.imagesDir(), fmt.Sprintf("gallery-%d", galleryID))
}

// Path is where the image lives on disk. Directories stay grouped by gallery
// so they are readable by a human; the filename is the uuid so uploads never
// collide and the public URL is unguessable.
func (s *ImageService) Path(img Image) string {
	return filepath.Join(s.galleryDir(img.GalleryID), img.UUID+img.Ext)
}

func (s *ImageService) imageContentTypes() []string {
	return []string{"image/png", "image/jpeg", "image/gif"}
}

func (s *ImageService) extensions() []string {
	return []string{".png", ".jpg", ".jpeg", ".gif"}
}
