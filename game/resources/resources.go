package resources

import (
	"bytes"
	"embed"
	"errors"
	"image"
	"io"
	"os"
	"strings"

	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/spf13/viper"
	"github.com/tinne26/etxt"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"

	log "github.com/sirupsen/logrus"
)

const SampleRate = 44100

var (
	UserConfigFile string
	UserKeymapFile string

	//go:embed audio fonts maps menu missions shaders sprites textures units weapons
	embedded embed.FS
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.SetEnvPrefix("pixelmek")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	userConfigPath, _ := os.UserHomeDir()
	if userConfigPath == "" {
		userConfigPath = "./"
	}
	userConfigPath += "/.pixelmek-3d"
	UserConfigFile = userConfigPath + "/config.json"
	UserKeymapFile = userConfigPath + "/keymap.json"

	viper.AddConfigPath(userConfigPath)

	err := viper.ReadInConfig()
	if err != nil {
		log.Error(err)
	}
}

func NewImageFromFile(path string) (*ebiten.Image, image.Image, error) {
	f, err := FileAt(path)
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
	b, err := ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return etxt.ParseFontBytes(b)
}

func LoadFont(path string, size float64) (font.Face, error) {
	fontData, err := ReadFile(path)
	if err != nil {
		return nil, err
	}

	switch {
	case strings.HasSuffix(path, ".otf"):
		otfFont, err := opentype.Parse(fontData)
		if err != nil {
			return nil, err
		}

		return opentype.NewFace(otfFont, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		})

	case strings.HasSuffix(path, ".ttf"):
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

func NewAudioStreamFromFile(path string) (io.ReadSeeker, int64, error) {
	audioBytes, err := ReadFile(path)
	if err != nil || audioBytes == nil {
		return nil, 0, err
	}
	return NewAudioStream(audioBytes, path)
}

func NewAudioStream(audioBytes []byte, path string) (io.ReadSeeker, int64, error) {
	reader := bytes.NewReader(audioBytes)

	var err error
	switch {
	case strings.HasSuffix(path, ".mp3"):
		stream, err := mp3.DecodeWithSampleRate(SampleRate, reader)
		return stream, stream.Length(), err
	case strings.HasSuffix(path, ".ogg"):
		stream, err := vorbis.DecodeWithSampleRate(SampleRate, reader)
		return stream, stream.Length(), err
	default:
		err = errors.New("unhandled audio extension for " + path)
	}

	return nil, 0, err

}

func NewShaderFromFile(path string) (*ebiten.Shader, error) {
	shaderData, err := ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ebiten.NewShader(shaderData)
}
