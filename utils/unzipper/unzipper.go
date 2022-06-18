package unzipper

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/brawaru/marct/utils"
	"github.com/brawaru/marct/utils/orderedmap"
	"github.com/brawaru/marct/validfile"
)

// sanePath returns either a path relative to the destination directory, or an error if it escapes the boundaries of
// the destination directory.
func sanePath(name string, dest string) (string, error) {
	dp := filepath.Join(dest, filepath.FromSlash(name))
	if !strings.HasPrefix(dp, filepath.Clean(dest)+string(os.PathSeparator)) {
		return "", fmt.Errorf("illegal path: %s", name)
	}
	return dp, nil
}

// FIXME: instead of processing functions having Unzipper instance passed to them, return more safe interface that,
//  for example, won't allow to close the extractor or change entries map (read-only).

// ErrSkip is a special error that can be returned from EntryProcessor to skip the extraction of an entry.
var ErrSkip = errors.New("skip")

// AbortErr is a special error that can be returned to abort the extraction of an entry. It's prominent use is during
// the validation of the file to signal critical errors unrelated to the validity of the file.
type AbortErr struct {
	Err error // Underlying error.
}

func (e *AbortErr) Error() string {
	return fmt.Sprintf("abort: %s", e.Err.Error())
}

func (e *AbortErr) Uwrap() error {
	return e.Err
}

func (e *AbortErr) Is(err error) bool {
	t, ok := err.(*AbortErr)
	return ok && errors.Is(t.Err, e.Err)
}

// EntryProcessor is a function that is called for each file in the archive before it is extracted, the arguments are
// as follows: name - name of the file in the archive, f - the file in the archive, dest - the destination (where the
// file is extracted). The function may process file in any desired way or return SkipErr to skip its extraction.
type EntryProcessor func(e *Unzipper, name string, f *zip.File, dest string) error

// DuplicateResolver is a function that is called when two files with the same name are found in the archive. The
// function is called with the name of the file, the first file and the second file. The function may return the file
// to use or an error.
type DuplicateResolver func(e *Unzipper, name string, a *zip.File, b *zip.File) (*zip.File, error)

// FileValidator is a function that is called for each extracted entry together with destination and entry
// itself, it checks whether the extracted file is valid or not. If the file is invalid, it should return an error,
// otherwise nil is expected. This function is never called for directories.
type FileValidator func(e *Unzipper, name string, dst string, f *os.File) error

type ExtractorOptions struct {
	EntryProcessor    EntryProcessor    // Function to process file before extraction.
	DuplicateResolver DuplicateResolver // Function to resolve duplicate files.
	FileValidator     FileValidator     // Function to validate existing or extracted files.
}

type Option func(*ExtractorOptions)

func WithEntryProcessor(f EntryProcessor) Option {
	return func(o *ExtractorOptions) {
		o.EntryProcessor = f
	}
}

func WithDuplicateResolver(f DuplicateResolver) Option {
	return func(o *ExtractorOptions) {
		o.DuplicateResolver = f
	}
}

// WithErrOnDuplicates returns an option that sets the DuplicateResolveFunc to a function that returns an error when
// two files with the same name are found in the archive.
func WithErrOnDuplicates() Option {
	return func(o *ExtractorOptions) {
		o.DuplicateResolver = func(_ *Unzipper, name string, a *zip.File, b *zip.File) (*zip.File, error) {
			return nil, fmt.Errorf("duplicate file: %s", name)
		}
	}
}

// WithFileValidator returns an option that sets the FileValidator to the given function.
func WithFileValidator(f FileValidator) Option {
	return func(o *ExtractorOptions) {
		o.FileValidator = f
	}
}

type Unzipper struct {
	reader  *zip.ReadCloser                   // Opened zip file.
	options *ExtractorOptions                 // Extractor options.
	Entries orderedmap.Map[string, *zip.File] // Map of files in the archive.
}

func (e *Unzipper) indexFiles() error {
	m := orderedmap.New[string, *zip.File]()

	for _, f := range e.reader.File {
		if m.HasKey(f.Name) {
			a := m.Get(f.Name)

			if e.options.DuplicateResolver != nil {
				chosen, err := e.options.DuplicateResolver(e, f.Name, a, f)
				if err != nil {
					return err
				}
				f = chosen
			} else {
				f = nil
			}

			if f == nil {
				continue
			}
		}

		m.Put(f.Name, f)
	}

	e.Entries = m

	return nil
}

func (e *Unzipper) init(name string) error {
	r, err := zip.OpenReader(name)
	if err != nil {
		return fmt.Errorf("open %q: %w", name, err)
	}
	e.reader = r

	if err := e.indexFiles(); err != nil {
		return fmt.Errorf("index files: %w", err)
	}

	return nil
}

func (e *Unzipper) Close() error {
	return e.reader.Close()
}

func NewExtractor(name string, opts ...Option) (*Unzipper, error) {
	e := &Unzipper{
		options: &ExtractorOptions{},
	}

	for _, opt := range opts {
		opt(e.options)
	}

	err := e.init(name)

	return e, err
}

func (e *Unzipper) writeEntryDst(f *zip.File, dst string) error {
	// If it's a directory, create it and return.
	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(dst, f.Mode()); err != nil {
			return fmt.Errorf("mkdir %q: %w", dst, err)
		}

		return nil
	}

	// Create all parent directories.
	fDir := filepath.Dir(dst)
	if exists, err := validfile.DirExists(fDir); err != nil {
		return fmt.Errorf("check dir %q: %w", fDir, err)
	} else if !exists {
		dirMode := fs.FileMode(0777)

		dirName := path.Dir(f.Name)
		if e.Entries.HasKey(dirName) {
			dirMode = e.Entries.Get(dirName).FileInfo().Mode()
		}

		if err := os.MkdirAll(fDir, dirMode); err != nil {
			return fmt.Errorf("mkdir %q: %w", fDir, err)
		}
	}

	if exists, err := validfile.FileExists(dst); err == nil {
		if exists && e.options.FileValidator != nil {
			if err := e.options.FileValidator(e, f.Name, dst, nil); err == nil {
				return nil // the file is valid
			} else if errors.Is(err, &AbortErr{}) {
				return err
			}
		}
	} else {
		return fmt.Errorf("exists %q: %w", dst, err)
	}

	outputFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return fmt.Errorf("open %q: %w", dst, err)
	}
	defer utils.DClose(outputFile)

	fileReader, err := f.Open()
	if err != nil {
		return fmt.Errorf("read entry: %w", err)
	}
	defer utils.DClose(fileReader)

	if _, err := io.Copy(outputFile, fileReader); err != nil {
		return fmt.Errorf("write %q: %w", dst, err)
	}

	return outputFile.Sync()
}

func (e *Unzipper) extractEntry(f *zip.File, dest string) error {
	dst, err := sanePath(f.Name, dest)
	if err != nil {
		return err
	}

	if err := e.writeEntryDst(f, dst); err != nil {
		return err
	}

	if err := e.options.FileValidator(e, f.Name, dst, nil); err != nil {
		return fmt.Errorf("post-validate %q: %w", dst, err)
	}

	return nil
}

func (e *Unzipper) Unzip(dest string) error {
	for _, k := range e.Entries.Keys() {
		v := e.Entries.Get(k)

		if e.options.EntryProcessor != nil {
			if err := e.options.EntryProcessor(e, k, v, dest); err != nil {
				if errors.Is(err, ErrSkip) {
					continue
				}

				return fmt.Errorf("process file %q: %w", k, err)
			}
		}

		if err := e.extractEntry(v, dest); err != nil {
			return fmt.Errorf("extract %q: %w", k, err)
		}
	}

	return nil
}

func Unzip(name string, dest string, opts ...Option) error {
	e, err := NewExtractor(name, opts...)
	if err != nil {
		return err
	}
	defer utils.DClose(e)

	return e.Unzip(dest)
}
