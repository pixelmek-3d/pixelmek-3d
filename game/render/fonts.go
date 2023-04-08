package render

import (
	"embed"
	"path/filepath"

	"github.com/tinne26/etxt"
	"github.com/tinne26/etxt/ecache"
)

type FontHandler struct {
	HUDFont  *Font
	embedded embed.FS
}

type Font struct {
	*etxt.Font
	FontCache *ecache.DefaultCache
	FontName  string
	FontPath  string
}

func NewFontHandler(embedded embed.FS) *FontHandler {
	f := &FontHandler{embedded: embedded}
	return f
}

func (f *FontHandler) LoadFont(fontFile string) (*Font, error) {
	fontPath := filepath.Join("resources", "fonts", fontFile)
	font, fontName, err := etxt.ParseEmbedFontFrom(filepath.ToSlash(fontPath), f.embedded)
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
