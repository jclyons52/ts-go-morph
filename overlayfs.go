package tsmorph

import (
	"io/fs"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/tspath"
	"github.com/jclyons52/ts-go-morph/third_party/typescript-go/ts/vfs"
)

// overlayFS is a vfs.FS that layers in-memory file contents over a base FS.
// All unsaved edits made through a Project live in the overlay, so the
// compiler always sees current text without touching disk.
type overlayFS struct {
	base vfs.FS

	mu       sync.RWMutex
	overlays map[string]string // canonical path -> text
	original map[string]string // canonical path -> path with original casing
}

func newOverlayFS(base vfs.FS) *overlayFS {
	return &overlayFS{
		base:     base,
		overlays: map[string]string{},
		original: map[string]string{},
	}
}

func (o *overlayFS) key(path string) string {
	return tspath.GetCanonicalFileName(tspath.NormalizePath(path), o.base.UseCaseSensitiveFileNames())
}

// setOverlay puts text into the in-memory overlay.
func (o *overlayFS) setOverlay(path, text string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	k := o.key(path)
	o.overlays[k] = text
	o.original[k] = tspath.NormalizePath(path)
}

// hasOverlay reports whether path has unsaved in-memory content.
func (o *overlayFS) hasOverlay(path string) bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	_, ok := o.overlays[o.key(path)]
	return ok
}

// flush writes all overlaid files to the base FS and clears the overlay.
func (o *overlayFS) flush() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	for k, text := range o.overlays {
		if err := o.base.WriteFile(o.original[k], text); err != nil {
			return err
		}
	}
	clear(o.overlays)
	clear(o.original)
	return nil
}

// flushFile writes a single overlaid file to the base FS and clears its
// overlay entry. Returns false if the file has no overlay.
func (o *overlayFS) flushFile(path string) (bool, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	k := o.key(path)
	text, ok := o.overlays[k]
	if !ok {
		return false, nil
	}
	if err := o.base.WriteFile(o.original[k], text); err != nil {
		return true, err
	}
	delete(o.overlays, k)
	delete(o.original, k)
	return true, nil
}

// isDirty reports whether any file has unsaved changes.
func (o *overlayFS) isDirty() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return len(o.overlays) > 0
}

func (o *overlayFS) UseCaseSensitiveFileNames() bool {
	return o.base.UseCaseSensitiveFileNames()
}

func (o *overlayFS) FileExists(path string) bool {
	if o.hasOverlay(path) {
		return true
	}
	return o.base.FileExists(path)
}

func (o *overlayFS) ReadFile(path string) (string, bool) {
	o.mu.RLock()
	if text, ok := o.overlays[o.key(path)]; ok {
		o.mu.RUnlock()
		return text, true
	}
	o.mu.RUnlock()
	return o.base.ReadFile(path)
}

func (o *overlayFS) WriteFile(path string, data string) error {
	// Writes through the public FS interface go to the overlay; they only
	// reach disk on flush (Save).
	o.setOverlay(path, data)
	return nil
}

func (o *overlayFS) AppendFile(path string, data string) error {
	contents, _ := o.ReadFile(path)
	o.setOverlay(path, contents+data)
	return nil
}

func (o *overlayFS) Remove(path string) error {
	o.mu.Lock()
	k := o.key(path)
	_, had := o.overlays[k]
	delete(o.overlays, k)
	delete(o.original, k)
	o.mu.Unlock()
	if had {
		return nil
	}
	return o.base.Remove(path)
}

func (o *overlayFS) Chtimes(path string, aTime time.Time, mTime time.Time) error {
	if o.hasOverlay(path) {
		return nil // in-memory files have no meaningful timestamps
	}
	return o.base.Chtimes(path, aTime, mTime)
}

func (o *overlayFS) DirectoryExists(path string) bool {
	if o.base.DirectoryExists(path) {
		return true
	}
	// A directory "exists" if any overlaid file lives beneath it.
	prefix := o.key(tspath.RemoveTrailingDirectorySeparator(path)) + "/"
	o.mu.RLock()
	defer o.mu.RUnlock()
	for k := range o.overlays {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}

func (o *overlayFS) GetAccessibleEntries(path string) vfs.Entries {
	entries := o.base.GetAccessibleEntries(path)
	prefix := tspath.RemoveTrailingDirectorySeparator(tspath.NormalizePath(path)) + "/"
	o.mu.RLock()
	defer o.mu.RUnlock()
	for _, orig := range o.original {
		if !strings.HasPrefix(orig, prefix) {
			continue
		}
		rest := orig[len(prefix):]
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			dir := rest[:i]
			if !slices.Contains(entries.Directories, dir) {
				entries.Directories = append(entries.Directories, dir)
			}
		} else if !slices.Contains(entries.Files, rest) {
			entries.Files = append(entries.Files, rest)
		}
	}
	slices.Sort(entries.Files)
	slices.Sort(entries.Directories)
	return entries
}

func (o *overlayFS) Stat(path string) vfs.FileInfo {
	if o.hasOverlay(path) {
		text, _ := o.ReadFile(path)
		return overlayFileInfo{name: tspath.GetBaseFileName(path), size: int64(len(text))}
	}
	return o.base.Stat(path)
}

func (o *overlayFS) WalkDir(root string, walkFn vfs.WalkDirFunc) error {
	// Overlay-only files are not visited; the compiler's directory walks are
	// for module resolution and tsconfig includes, which consult FileExists /
	// GetAccessibleEntries directly.
	return o.base.WalkDir(root, walkFn)
}

func (o *overlayFS) Realpath(path string) string {
	if o.hasOverlay(path) {
		return tspath.NormalizePath(path)
	}
	return o.base.Realpath(path)
}

// overlayFileInfo is a minimal fs.FileInfo for in-memory files.
type overlayFileInfo struct {
	name string
	size int64
}

func (fi overlayFileInfo) Name() string       { return fi.name }
func (fi overlayFileInfo) Size() int64        { return fi.size }
func (fi overlayFileInfo) Mode() fs.FileMode  { return 0o644 }
func (fi overlayFileInfo) ModTime() time.Time { return time.Time{} }
func (fi overlayFileInfo) IsDir() bool        { return false }
func (fi overlayFileInfo) Sys() any           { return nil }
