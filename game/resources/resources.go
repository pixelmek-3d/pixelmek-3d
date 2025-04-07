package resources

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	efont "github.com/tinne26/etxt/font"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"
	v "github.com/spf13/viper"
)

const SampleRate = 44100

var (
	Viper          *v.Viper
	UserConfigFile string
	UserKeymapFile string

	CrosshairsSheet *CrosshairsSheetConfig

	//go:embed ai audio fonts icons maps menu missions shaders sprites textures all:units all:weapons
	embedded embed.FS
)

type CrosshairsSheetConfig struct {
	Columns int `yaml:"columns" validate:"gt=0"`
	Rows    int `yaml:"rows" validate:"gt=0"`
}

func InitConfig() {
	Viper = v.New()
	Viper.SetConfigName("config")
	Viper.SetConfigType("json")
	Viper.SetEnvPrefix("pixelmek")
	Viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	Viper.AutomaticEnv()

	userConfigPath, _ := os.UserHomeDir()
	if userConfigPath == "" {
		userConfigPath = "./"
	}
	userConfigPath += "/.pixelmek-3d"
	UserConfigFile = userConfigPath + "/config.json"
	UserKeymapFile = userConfigPath + "/keymap.json"

	Viper.AddConfigPath(userConfigPath)

	err := Viper.ReadInConfig()
	if err != nil {
		log.Error(err)
	}

	initConfigFS()
}

func InitResources() {
	// initialize resource file system handler
	initFS()

	// load crosshairs sheet configuration
	crosshairsConfigFile := path.Join("sprites", "hud", "crosshairs_sheet.yaml")

	// TODO: refactor the following into generic use function
	fileContent, err := ReadFile(crosshairsConfigFile)
	if err != nil {
		log.Fatal(fmt.Errorf("[%s] %s", crosshairsConfigFile, err.Error()))
	}

	CrosshairsSheet = &CrosshairsSheetConfig{}
	err = yaml.Unmarshal(fileContent, CrosshairsSheet)
	if err != nil {
		log.Fatal(fmt.Errorf("[%s] %s", crosshairsConfigFile, err.Error()))
	}

	v := validator.New()
	err = v.Struct(CrosshairsSheet)
	if err != nil {
		log.Fatal(fmt.Errorf("[%s] %s", crosshairsConfigFile, err.Error()))
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
	return efont.ParseFromBytes(b)
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

func BaseNameWithoutExtension(file string) string {
	return strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
}

func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Invalid, reflect.Pointer, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
