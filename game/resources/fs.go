package resources

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/spf13/viper"

	tarfs "github.com/nlepage/go-tarfs"
	log "github.com/sirupsen/logrus"
)

const (
	modsPath = "mods"
)

var (
	hasLocalResources bool = false
	fsPathMap         map[string]map[string]*fsResource
)

type fsResource struct {
	entry fs.DirEntry
	fs    fs.FS
}

func initConfigFS() {
	ignoreLocal := viper.GetBool("ignoreLocalResources")
	if !ignoreLocal {
		info, err := os.Stat(modsPath)
		hasLocalResources = !errors.Is(err, fs.ErrNotExist) && info != nil && info.IsDir()
	}
}

func initFS() {
	fsPathMap = make(map[string]map[string]*fsResource)

	// load embedFS entries into fsPathMap
	_, err := embedded.Open(".")
	if err != nil {
		panic(err)
	}
	fs.WalkDir(embedded, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Errorf("error walking mods file %s", err)
			return nil
		}

		_storeFsResource(p, d, embedded)
		return nil
	})

	if !hasLocalResources {
		return
	}

	// load mods/*.tar and walk their paths to store the FS (FileSystem) instance for each path
	// * last file resource entry by name "wins"
	modsDir, err := os.Open(modsPath)
	if err == nil {
		modFiles, err := modsDir.ReadDir(0)
		if err == nil && len(modFiles) > 0 {
			for _, t := range modFiles {
				tName := t.Name()
				tPath := filepath.Join(modsDir.Name(), tName)
				if filepath.Ext(tName) != ".tar" {
					continue
				}

				// walk the files in the archive
				tfs, err := _loadTarFS(tPath)
				if err != nil {
					log.Errorf("error loading mods file %s", err)
					continue
				}
				log.Debugf("found mods file %s", tPath)
				fs.WalkDir(tfs, ".", func(p string, d fs.DirEntry, err error) error {
					if err != nil {
						log.Errorf("error walking mods file %s", err)
						return nil
					}
					log.Debugf("[%s] %s", tPath, p)
					_storeFsResource(p, d, tfs)
					return nil
				})
			}
		}
	}
}

func _storeFsResource(p string, d fs.DirEntry, _fs fs.FS) {
	if p == "." {
		return
	}

	fName := d.Name()
	fParent := path.Dir(p)

	fsr := &fsResource{
		entry: d,
		fs:    _fs,
	}

	fsPathSub, ok := fsPathMap[fParent]
	if !ok {
		fsPathSub = make(map[string]*fsResource)
		fsPathMap[fParent] = fsPathSub
	}

	fsPathSub[fName] = fsr
}

func _loadTarFS(path string) (fs.FS, error) {
	tf, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	tfs, err := tarfs.New(tf)
	if err != nil {
		tf.Close()
		return nil, err
	}

	return tfs, nil
}

func _fsForPath(fPath string) (fs.FS, error) {
	pDir := path.Dir(fPath)
	fsPathSub, ok := fsPathMap[pDir]
	if !ok {
		return nil, fmt.Errorf("directory not found for %s", fPath)
	}

	pBase := path.Base(fPath)
	fsr, ok := fsPathSub[pBase]
	if !ok {
		return nil, fmt.Errorf("file not found for %s", fPath)
	}

	return fsr.fs, nil
}

func FileAt(path string) (fs.File, error) {
	tfs, err := _fsForPath(path)
	if err != nil {
		return nil, err
	}
	return tfs.Open(path)
}

func ReadFile(path string) ([]byte, error) {
	f, err := FileAt(path)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(f)
}

func ReadDir(dir string, recurseDirs bool) ([]DirEntryWithParent, error) {
	return readDir(dir, recurseDirs, "")
}

// DirEntryWithParent is an fs.DirEntry with relative parent path prefixed
type DirEntryWithParent struct {
	fs.DirEntry
	parent string
}

func newDirEntryWithParent(f fs.DirEntry, parent string) DirEntryWithParent {
	return DirEntryWithParent{
		DirEntry: f,
		parent:   parent,
	}
}

// Name returns the name of the file entry (or subdirectory).
// This name contains a relative parent path prefixed starting from the point where ReadDir was called.
func (f DirEntryWithParent) Name() string {
	return path.Join(f.parent, f.DirEntry.Name())
}

// Parent returns the relative parent path prefix
func (f DirEntryWithParent) Parent() string {
	return f.parent
}

func readDir(dir string, recurseDirs bool, recurseParent string) ([]DirEntryWithParent, error) {
	fsPathSub, ok := fsPathMap[dir]
	if !ok {
		return nil, fmt.Errorf("directory not found for %s", dir)
	}

	// sort filename keys
	keys := make([]string, 0, len(fsPathSub))
	for k := range fsPathSub {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var entries []DirEntryWithParent = make([]DirEntryWithParent, 0, len(fsPathSub))
	for _, k := range keys {
		fsr := fsPathSub[k]
		entries = append(entries, newDirEntryWithParent(fsr.entry, recurseParent))

		if fsr.entry.IsDir() && recurseDirs {
			subEntries, err := readDir(path.Join(dir, fsr.entry.Name()), true, path.Join(recurseParent, fsr.entry.Name()))
			if err != nil {
				return nil, err
			}
			entries = append(entries, subEntries...)
		}
	}
	return entries, nil
}
