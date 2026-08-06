package local

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wrallee/review-focus-mcp/internal/domain"
)

type Store struct{ Root string }

func New(root string) *Store { return &Store{Root: root} }

func (s *Store) SaveAnalysis(a domain.Analysis) error {
	return s.write("analyses", key(a.Repository, a.Number, a.HeadSHA)+".json", a)
}
func (s *Store) LoadAnalysis(repository string, number int, headSHA string) (*domain.Analysis, error) {
	var a domain.Analysis
	if err := s.read("analyses", key(repository, number, headSHA)+".json", &a); err != nil {
		return nil, err
	}
	return &a, nil
}
func (s *Store) SaveDraft(d domain.ReviewDraft) error {
	return s.write("drafts", key(d.Repository, d.Number, d.HeadSHA)+".json", d)
}
func (s *Store) LoadDraft(repository string, number int, headSHA string) (*domain.ReviewDraft, error) {
	var d domain.ReviewDraft
	if err := s.read("drafts", key(repository, number, headSHA)+".json", &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Store) FindLatestDraft(repository string, number int) (*domain.ReviewDraft, error) {
	dir := filepath.Join(s.Root, "drafts")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	prefix := keyPrefix(repository, number)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) {
			var d domain.ReviewDraft
			if err := s.read("drafts", e.Name(), &d); err == nil {
				return &d, nil
			}
		}
	}
	return nil, os.ErrNotExist
}

func (s *Store) write(bucket, name string, v any) error {
	dir := filepath.Join(s.Root, bucket)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, name+".tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dir, name))
}
func (s *Store) read(bucket, name string, v any) error {
	b, err := os.ReadFile(filepath.Join(s.Root, bucket, name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.ErrNotExist
		}
		return err
	}
	return json.Unmarshal(b, v)
}
func key(repository string, number int, headSHA string) string {
	return fmt.Sprintf("%s__%d__%s", sanitize(repository), number, headSHA)
}
func keyPrefix(repository string, number int) string {
	return fmt.Sprintf("%s__%d__", sanitize(repository), number)
}
func sanitize(v string) string { return strings.NewReplacer("/", "_", "\\", "_", ":", "_").Replace(v) }
