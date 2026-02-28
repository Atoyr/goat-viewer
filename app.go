package main

import (
	"archive/zip"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// App struct
type App struct {
	ctx       context.Context
	zr        *zip.ReadCloser
	pageIdx   []int
	pageNames []string // decoded display names (dir/file)
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ChooseAndOpenArchive shows a file dialog and opens the selected archive.
// Returns the sorted list of page filenames.
func (a *App) ChooseAndOpenArchive() ([]string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Comic ZIP/CBZ",
		Filters: []runtime.FileFilter{{
			DisplayName: "Comic archives (*.zip, *.cbz)",
			Pattern:     "*.zip;*.cbz",
		}},
	})
	if err != nil || path == "" {
		return nil, err
	}
	if err := a.OpenArchive(path); err != nil {
		return nil, err
	}
	return a.ListPages(), nil
}

// OpenArchive opens a zip/cbz file by path.
func (a *App) OpenArchive(path string) error {
	// Close any previous archive
	if a.zr != nil {
		_ = a.zr.Close()
		a.zr = nil
		a.pageIdx = nil
		a.pageNames = nil
	}
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	a.zr = zr

	// Collect image file indices
	var idxs []int
	for i, f := range zr.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if isImage(f.Name) {
			idxs = append(idxs, i)
		}
	}
	// Sort by directory first, then by filename, both using natural order.
	// ZIP paths always use forward slashes.
	sort.Slice(idxs, func(i, j int) bool {
		ni := zr.File[idxs[i]].Name
		nj := zr.File[idxs[j]].Name
		slashI := strings.LastIndex(ni, "/")
		slashJ := strings.LastIndex(nj, "/")
		dirI, baseI := ni[:slashI+1], ni[slashI+1:]
		dirJ, baseJ := nj[:slashJ+1], nj[slashJ+1:]
		if dirI != dirJ {
			return naturalLess(dirI, dirJ)
		}
		return naturalLess(baseI, baseJ)
	})
	a.pageIdx = idxs
	a.pageNames = make([]string, len(idxs))
	for i, idx := range idxs {
		a.pageNames[i] = decodeName(zr.File[idx])
	}
	return nil
}

// ListPages returns the sorted page names as "dir/file" (decoded).
func (a *App) ListPages() []string {
	if a.zr == nil {
		return nil
	}
	out := make([]string, len(a.pageNames))
	copy(out, a.pageNames)
	return out
}

// GetPageDataURL returns a data URL (image/*, base64) for the given page index.
func (a *App) GetPageDataURL(index int) (string, error) {
	if a.zr == nil || index < 0 || index >= len(a.pageIdx) {
		return "", fmt.Errorf("invalid page index")
	}
	f := a.zr.File[a.pageIdx[index]]
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	mime := mimeFromName(f.Name)
	enc := base64.StdEncoding.EncodeToString(b)
	return fmt.Sprintf("data:%s;base64,%s", mime, enc), nil
}

// CloseArchive releases the current archive.
func (a *App) CloseArchive() error {
	if a.zr != nil {
		err := a.zr.Close()
		a.zr = nil
		a.pageIdx = nil
		a.pageNames = nil
		return err
	}
	return nil
}

// Helpers

// decodeName returns a human-readable path from a ZIP file entry.
// ZIP entries created on Japanese Windows use Shift-JIS when the UTF-8 flag is absent.
func decodeName(f *zip.File) string {
	name := f.Name
	if !f.NonUTF8 && utf8.ValidString(name) {
		return name
	}
	decoded, _, err := transform.String(japanese.ShiftJIS.NewDecoder(), name)
	if err != nil {
		return name
	}
	return decoded
}

func isImage(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
		return true
	default:
		return false
	}
}

func mimeFromName(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	default:
		return "application/octet-stream"
	}
}

var numChunk = regexp.MustCompile(`\d+|\D+`)

// naturalLess compares strings using natural numeric ordering.
func naturalLess(a, b string) bool {
	aa := strings.ToLower(a)
	bb := strings.ToLower(b)
	as := numChunk.FindAllString(aa, -1)
	bs := numChunk.FindAllString(bb, -1)
	for i := 0; i < len(as) && i < len(bs); i++ {
		sa, sb := as[i], bs[i]
		// If both numeric, compare numerically by length then lexicographically
		if isDigit(sa[0]) && isDigit(sb[0]) {
			// strip leading zeros for length compare
			tsa := strings.TrimLeft(sa, "0")
			tsb := strings.TrimLeft(sb, "0")
			if len(tsa) != len(tsb) {
				return len(tsa) < len(tsb)
			}
			if tsa != tsb {
				return tsa < tsb
			}
			// fall through if equal (including all zeros)
		} else if sa != sb {
			return sa < sb
		}
	}
	return len(as) < len(bs)
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }
