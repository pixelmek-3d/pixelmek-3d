package resources

import (
	"embed"
	"image"
	"io/fs"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font/sfnt"
)

//go:embed fonts maps missions sprites textures units weapons
var embedded embed.FS

func NewImageFromFile(path string) (*ebiten.Image, image.Image, error) {
	f, err := embedded.Open(filepath.ToSlash(path))
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	eb, im, err := ebitenutil.NewImageFromReader(f)
	return eb, im, err
}

func NewFontFromFile(path string) (*sfnt.Font, string, error) {
	return etxt.ParseEmbedFontFrom(filepath.ToSlash(path), embedded)
}

func FilesInPath(path string) ([]fs.DirEntry, error) {
	return embedded.ReadDir(filepath.ToSlash(path))
}

func ReadFile(name string) ([]byte, error) {
	return embedded.ReadFile(filepath.ToSlash(name))
}

func GetAllFilenames(efs embed.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		files = append(files, path)

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
