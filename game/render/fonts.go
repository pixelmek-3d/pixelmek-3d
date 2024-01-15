package render

import (
	"path"

	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/ecache"
)

type FontHandler struct {
	HUDFont *Font
}

type Font struct {
	*etxt.Font
	FontCache *ecache.DefaultCache
	FontName  string
	FontPath  string
}

func NewFontHandler(hudFont string) (*FontHandler, error) {
	var err error
	f := &FontHandler{}
	f.HUDFont, err = f.LoadFont(hudFont)
	return f, err
}

func (f *FontHandler) LoadFont(fontFile string) (*Font, error) {
	fontPath := path.Join("fonts", fontFile)
	font, fontName, err := resources.NewFontFromFile(fontPath)
	if err != nil {
		return nil, err
	}

	// create 10MB cache
	fontCache := etxt.NewDefaultCache(10 * 1024 * 1024)

	return &Font{
		Font:      font,
		FontCache: fontCache,
		FontName:  fontName,
		FontPath:  fontPath,
	}, nil
}
