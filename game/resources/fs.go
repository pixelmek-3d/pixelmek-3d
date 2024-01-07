package resources

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

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

func init() {
	ignoreLocal := viper.GetBool("ignoreLocalResources")
	if !ignoreLocal {
		info, err := os.Stat(modsPath)
		hasLocalResources = !errors.Is(err, fs.ErrNotExist) && info != nil && info.IsDir()
	}
}

func InitFS() {
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
	fParent := filepath.Dir(p)

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

func _fsForPath(path string) (fs.FS, error) {
	pDir := filepath.Dir(path)
	fsPathSub, ok := fsPathMap[pDir]
	if !ok {
		return nil, fmt.Errorf("directory not found for %s", path)
	}

	pBase := filepath.Base(path)
	fsr, ok := fsPathSub[pBase]
	if !ok {
		return nil, fmt.Errorf("file not found for %s", path)
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

func ReadDir(path string) ([]fs.DirEntry, error) {
	fsPathSub, ok := fsPathMap[path]
	if !ok {
		return nil, fmt.Errorf("directory not found for %s", path)
	}

	var entries []fs.DirEntry = make([]fs.DirEntry, 0, len(fsPathSub))
	for _, fsr := range fsPathSub {
		entries = append(entries, fsr.entry)
	}
	return entries, nil
}
