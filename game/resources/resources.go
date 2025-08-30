package resources

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	efont "github.com/tinne26/etxt/font"
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

	TexWidth    int
	imageByPath = make(map[string]*ebiten.Image)
	rgbaByPath  = make(map[string]*image.RGBA)

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

func InitResources(texWidth int) {
	// initialize resource file system handler
	initFS()

	TexWidth = texWidth

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

func LoadFont(path string, size float64) (text.Face, error) {
	fontData, err := ReadFile(path)
	if err != nil {
		return nil, err
	}

	ttfFont, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		return nil, fmt.Errorf("[%s] error loading font: %v", path, err)
	}

	return &text.GoTextFace{
		Source: ttfFont,
		Size:   size,
	}, nil
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

func GetRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	resourcePath := "textures"
	texFilePath := path.Join(resourcePath, texFile)
	if rgba, ok := rgbaByPath[texFilePath]; ok {
		return rgba
	}

	_, tex, err := NewImageFromFile(texFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, TexWidth, TexWidth))
		// convert into RGBA format
		for x := 0; x < TexWidth; x++ {
			for y := 0; y < TexWidth; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	if rgba != nil {
		rgbaByPath[resourcePath] = rgba
	}

	return rgba
}

func GetTextureFromFile(texFile string) *ebiten.Image {
	resourcePath := path.Join("textures", texFile)
	if eImg, ok := imageByPath[resourcePath]; ok {
		return eImg
	}

	eImg, _, err := NewImageFromFile(resourcePath)
	if err != nil {
		log.Fatal(err)
	}
	if eImg != nil {
		imageByPath[resourcePath] = eImg
	}
	return eImg
}

func GetSpriteFromFile(sFile string) *ebiten.Image {
	resourcePath := path.Join("sprites", sFile)
	if eImg, ok := imageByPath[resourcePath]; ok {
		return eImg
	}

	eImg, _, err := NewImageFromFile(resourcePath)
	if err != nil {
		log.Fatal(err)
	}
	if eImg != nil {
		imageByPath[resourcePath] = eImg
	}
	return eImg
}
