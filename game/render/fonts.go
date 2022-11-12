package render

import (
	"path/filepath"

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

func NewFontHandler() *FontHandler {
	f := &FontHandler{}
	return f
}

func (f *FontHandler) LoadFont(fontFile string) (*Font, error) {
	fontPath := filepath.Join("game", "resources", "fonts", fontFile)
	font, fontName, err := etxt.ParseFontFrom(fontPath)
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
