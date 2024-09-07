package filestorage

import (
	"encoding/json"
	"io"
	"os"

	"github.com/madatsci/urlshortener/internal/app/storage"
)

type (
	// Storage is an implementation of the URL storage which uses a file to save data on disk.
	Storage struct {
		filepath string
		urls     map[string]string
	}
	record struct {
		OriginalURL string `json:"original_url"`
		ShortURL    string `json:"short_url"`
	}
)

// New creates a new file storage.
func New(filepath string) (*Storage, error) {
	s := &Storage{
		filepath: filepath,
		urls:     make(map[string]string),
	}

	if err := s.load(); err != nil {
		return s, err
	}

	return s, nil
}

// Add adds a new URL with its slug to the storage.
func (s *Storage) Add(slug string, url string) error {
	s.urls[slug] = url

	file, err := os.OpenFile(s.filepath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	record := &record{
		ShortURL:    slug,
		OriginalURL: url,
	}

	return json.NewEncoder(file).Encode(record)
}

// Get retrieves a URL by its slug from the storage.
func (s *Storage) Get(slug string) (string, error) {
	url, ok := s.urls[slug]
	if !ok {
		return "", storage.ErrURLNotFound
	}

	return url, nil
}

// ListAll returns the full map of stored URLs.
func (s *Storage) ListAll() map[string]string {
	return s.urls
}

func (s *Storage) load() error {
	file, err := os.OpenFile(s.filepath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	urls := make(map[string]string)

	for {
		record := &record{}
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		urls[record.ShortURL] = record.OriginalURL
	}

	s.urls = urls
	return nil
}
