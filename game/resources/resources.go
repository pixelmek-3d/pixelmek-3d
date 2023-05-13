package resources

import (
	"embed"
	"errors"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"

	log "github.com/sirupsen/logrus"
)

var (
	//go:embed fonts maps menu missions sprites textures units weapons
	embedded          embed.FS
	hasLocalResources bool = false
)

func init() {
	info, err := os.Stat(filepath.Join("game", "resources"))
	hasLocalResources = !errors.Is(err, fs.ErrNotExist) && info != nil && info.IsDir()
}

func NewImageFromFile(path string) (*ebiten.Image, image.Image, error) {
	if hasLocalResources {
		// check for local override of file
		localPath, ok := localResourceCheck(path)
		if ok {
			log.Debugf("using local override %s", localPath)
			return ebitenutil.NewImageFromFile(localPath)
		}
	}

	f, err := embedded.Open(filepath.ToSlash(path))
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	eb, im, err := ebitenutil.NewImageFromReader(f)
	return eb, im, err
}

func NewScaledImageFromFile(path string, scale float64) (*ebiten.Image, image.Image, error) {
	eb, im, err := NewImageFromFile(path)
	if err != nil {
		return eb, im, err
	}

	if scale == 1.0 {
		return eb, im, err
	}

	op := &ebiten.DrawImageOptions{}
	op.Filter = ebiten.FilterNearest
	op.GeoM.Scale(scale, scale)

	scaledWidth, scaledHeight := float64(eb.Bounds().Dx())*scale, float64(eb.Bounds().Dy())*scale
	scaledImage := ebiten.NewImage(int(scaledWidth), int(scaledHeight))
	scaledImage.DrawImage(eb, op)

	return scaledImage, scaledImage, err
}

func NewFontFromFile(path string) (*sfnt.Font, string, error) {
	return etxt.ParseEmbedFontFrom(filepath.ToSlash(path), embedded)
}

func LoadFont(path string, size float64) (font.Face, error) {
	fontData, err := embedded.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(path, ".otf") {
		otfFont, err := opentype.Parse(fontData)
		if err != nil {
			return nil, err
		}

		return opentype.NewFace(otfFont, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		})

	} else if strings.HasSuffix(path, ".ttf") {
		ttfFont, err := truetype.Parse(fontData)
		if err != nil {
			return nil, err
		}

		return truetype.NewFace(ttfFont, &truetype.Options{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		}), nil
	}

	return nil, errors.New("unhandled font extension for " + path)
}

func FilesInPath(path string) ([]fs.DirEntry, error) {
	files := make([]fs.DirEntry, 0)

	embeddedFiles, err := embedded.ReadDir(filepath.ToSlash(path))
	if err == nil {
		files = append(files, embeddedFiles...)
	}

	if hasLocalResources {
		// check for additional/override local resources
		localPath, ok := localResourceCheck(path)
		if ok {
			localDir, err := os.Open(localPath)
			if err == nil {
				localFiles, err := localDir.ReadDir(0)
				if err == nil && len(localFiles) > 0 {
					log.Debugf("found local override files in %s", localPath)
					for _, l := range localFiles {
						found := false
						for _, f := range files {
							found = l.Name() == f.Name()
							if found {
								break
							}
						}
						if !found {
							log.Debugf("found local additional file %s", filepath.Join(localPath, l.Name()))
							files = append(files, l)
						}
					}
				}
			}
		}
	}

	return files, err
}

func ReadFile(path string) ([]byte, error) {
	if hasLocalResources {
		// check for local override of file
		localPath, ok := localResourceCheck(path)
		if ok {
			log.Debugf("using local override %s", localPath)
			return os.ReadFile(localPath)
		}
	}

	return embedded.ReadFile(filepath.ToSlash(path))
}

func GetEmbeddedFiles(efs embed.FS) (files []string, err error) {
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

func localResourceCheck(path string) (string, bool) {
	localPath := filepath.Join("game", "resources", path)
	_, err := os.Stat(localPath)
	return localPath, !errors.Is(err, fs.ErrNotExist)
}
