package game

import (
	"image/color"
	"strconv"

	"github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/pixelmek-3d/pixelmek-3d/game/resources"
)

const (
	backgroundColor = "131a22"

	textIdleColor     = "dff4ff"
	textDisabledColor = "5a7a91"

	labelIdleColor     = textIdleColor
	labelDisabledColor = textDisabledColor

	buttonIdleColor     = textIdleColor
	buttonDisabledColor = labelDisabledColor

	listSelectedBackground         = "4b687a"
	listDisabledSelectedBackground = "2a3944"

	listFocusedBackground = "2a3944"

	headerColor = textIdleColor

	toolTipColor = backgroundColor

	separatorColor = listDisabledSelectedBackground
)

const (
	fontFaceRegular = "fonts/pixeloid-sans.otf"
	fontFaceBold    = "fonts/pixeloid-bold.otf"
	fontFaceMono    = "fonts/pixeloid.otf"
)

type uiResources struct {
	fonts           *fonts
	background      *image.NineSlice
	backgroundColor color.Color
	separatorColor  color.Color

	text        *textResources
	button      *buttonResources
	label       *labelResources
	checkbox    *checkboxResources
	comboButton *comboButtonResources
	list        *listResources
	slider      *sliderResources
	panel       *panelResources
	tabBook     *tabBookResources
	header      *headerResources
	textArea    *textAreaResources
	toolTip     *toolTipResources
}

type textResources struct {
	idleColor     color.Color
	disabledColor color.Color
	face          text.Face
	titleFace     text.Face
	bigTitleFace  text.Face
	smallFace     text.Face
}

type buttonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    text.Face
	padding widget.Insets
}

type checkboxResources struct {
	image   *widget.ButtonImage
	graphic *widget.CheckboxGraphicImage
	spacing int
}

type labelResources struct {
	text *widget.LabelColor
	face text.Face
}

type comboButtonResources struct {
	image   *widget.ButtonImage
	text    *widget.ButtonTextColor
	face    text.Face
	graphic *widget.GraphicImage
	padding widget.Insets
}

type listResources struct {
	image        *widget.ScrollContainerImage
	track        *widget.SliderTrackImage
	trackPadding widget.Insets
	handle       *widget.ButtonImage
	handleSize   int
	face         text.Face
	entry        *widget.ListEntryColor
	entryPadding widget.Insets
}

type sliderResources struct {
	trackImage *widget.SliderTrackImage
	handle     *widget.ButtonImage
	handleSize int
}

type panelResources struct {
	image    *image.NineSlice
	filled   *image.NineSlice
	titleBar *image.NineSlice
	padding  widget.Insets
}

type tabBookResources struct {
	buttonFace    text.Face
	buttonText    *widget.ButtonTextColor
	buttonPadding widget.Insets
}

type headerResources struct {
	background *image.NineSlice
	padding    widget.Insets
	face       text.Face
	color      color.Color
}

type textAreaResources struct {
	image        *widget.ScrollContainerImage
	track        *widget.SliderTrackImage
	trackPadding widget.Insets
	handle       *widget.ButtonImage
	handleSize   int
	face         text.Face
	entryPadding widget.Insets
}

type toolTipResources struct {
	background *image.NineSlice
	padding    widget.Insets
	face       text.Face
	monoFace   text.Face
	color      color.Color
}

type fonts struct {
	scale        float64
	face         text.Face
	titleFace    text.Face
	bigTitleFace text.Face
	toolTipFace  text.Face
	toolTipMono  text.Face
}

func NewUIResources(fonts *fonts) (*uiResources, error) {
	background := image.NewNineSliceColor(hexToColorAlpha(backgroundColor, 96))

	button, err := newButtonResources(fonts)
	if err != nil {
		return nil, err
	}

	checkbox, err := newCheckboxResources(fonts)
	if err != nil {
		return nil, err
	}

	comboButton, err := newComboButtonResources(fonts)
	if err != nil {
		return nil, err
	}

	list, err := newListResources(fonts)
	if err != nil {
		return nil, err
	}

	slider, err := newSliderResources()
	if err != nil {
		return nil, err
	}

	panel, err := newPanelResources()
	if err != nil {
		return nil, err
	}

	tabBook, err := newTabBookResources(fonts)
	if err != nil {
		return nil, err
	}

	header, err := newHeaderResources(fonts)
	if err != nil {
		return nil, err
	}

	textArea, err := newTextAreaResources(fonts)
	if err != nil {
		return nil, err
	}

	toolTip, err := newToolTipResources(fonts)
	if err != nil {
		return nil, err
	}

	return &uiResources{
		fonts:           fonts,
		background:      background,
		backgroundColor: hexToColor(backgroundColor),
		separatorColor:  hexToColor(separatorColor),

		text: &textResources{
			idleColor:     hexToColor(textIdleColor),
			disabledColor: hexToColor(textDisabledColor),
			face:          fonts.face,
			titleFace:     fonts.titleFace,
			bigTitleFace:  fonts.bigTitleFace,
			smallFace:     fonts.toolTipFace,
		},

		button:      button,
		label:       newLabelResources(fonts),
		checkbox:    checkbox,
		comboButton: comboButton,
		list:        list,
		slider:      slider,
		panel:       panel,
		tabBook:     tabBook,
		header:      header,
		textArea:    textArea,
		toolTip:     toolTip,
	}, nil
}

func loadFonts(fontScale float64) (*fonts, error) {
	fontFace, err := resources.LoadFont(fontFaceRegular, 20.0*fontScale)
	if err != nil {
		return nil, err
	}

	titleFontFace, err := resources.LoadFont(fontFaceBold, 24.0*fontScale)
	if err != nil {
		return nil, err
	}

	bigTitleFontFace, err := resources.LoadFont(fontFaceBold, 28.0*fontScale)
	if err != nil {
		return nil, err
	}

	toolTipFace, err := resources.LoadFont(fontFaceRegular, 15.0*fontScale)
	if err != nil {
		return nil, err
	}

	toolTipMono, err := resources.LoadFont(fontFaceMono, 15.0*fontScale)
	if err != nil {
		return nil, err
	}

	return &fonts{
		scale:        fontScale,
		face:         fontFace,
		titleFace:    titleFontFace,
		bigTitleFace: bigTitleFontFace,
		toolTipFace:  toolTipFace,
		toolTipMono:  toolTipMono,
	}, nil
}

func loadGraphicImages(idle string, disabled string, scale float64) (*widget.GraphicImage, error) {
	idleImage, _, err := resources.NewScaledImageFromFile(idle, scale)
	if err != nil {
		return nil, err
	}

	var disabledImage *ebiten.Image
	if disabled != "" {
		disabledImage, _, err = resources.NewScaledImageFromFile(disabled, scale)
		if err != nil {
			return nil, err
		}
	}

	return &widget.GraphicImage{
		Idle:     idleImage,
		Disabled: disabledImage,
	}, nil
}

func loadImageNineSlice(path string, centerWidth int, centerHeight int, scale float64) (*image.NineSlice, error) {
	i, _, err := resources.NewScaledImageFromFile(path, scale)
	if err != nil {
		return nil, err
	}
	w, h := i.Bounds().Dx(), i.Bounds().Dy()
	return image.NewNineSlice(i,
			[3]int{(w - centerWidth) / 2, centerWidth, w - (w-centerWidth)/2 - centerWidth},
			[3]int{(h - centerHeight) / 2, centerHeight, h - (h-centerHeight)/2 - centerHeight}),
		nil
}

func centerHeightFromFontScale(fontScale float64) int {
	if fontScale > 1.0 {
		// value must be 1 when font scale goes over 1.0
		return 1
	}
	return 0
}

func resourceScaleFromFontScale(fontScale float64) float64 {
	if fontScale > 1.0 {
		// resource scale must be no higher than 1 when font scale goes over 1.0
		return 1.0
	}
	return fontScale
}

func newButtonResources(fonts *fonts) (*buttonResources, error) {
	cH := centerHeightFromFontScale(fonts.scale)
	rS := resourceScaleFromFontScale(fonts.scale)
	idle, err := loadImageNineSlice("menu/button-idle.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}

	hover, err := loadImageNineSlice("menu/button-hover.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}
	pressed_hover, err := loadImageNineSlice("menu/button-selected-hover.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}
	pressed, err := loadImageNineSlice("menu/button-pressed.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}

	disabled, err := loadImageNineSlice("menu/button-disabled.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}

	i := &widget.ButtonImage{
		Idle:         idle,
		Hover:        hover,
		Pressed:      pressed,
		PressedHover: pressed_hover,
		Disabled:     disabled,
	}

	return &buttonResources{
		image: i,

		text: &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonDisabledColor),
		},

		face: fonts.face,

		padding: widget.Insets{
			Left:  30,
			Right: 30,
		},
	}, nil
}

func newCheckboxResources(fonts *fonts) (*checkboxResources, error) {
	cH := centerHeightFromFontScale(fonts.scale)
	rS := resourceScaleFromFontScale(fonts.scale)
	idle, err := loadImageNineSlice("menu/checkbox-idle.png", 20, cH, rS)
	if err != nil {
		return nil, err
	}

	hover, err := loadImageNineSlice("menu/checkbox-hover.png", 20, cH, rS)
	if err != nil {
		return nil, err
	}

	disabled, err := loadImageNineSlice("menu/checkbox-disabled.png", 20, cH, rS)
	if err != nil {
		return nil, err
	}

	checked, err := loadGraphicImages("menu/checkbox-checked-idle.png", "menu/checkbox-checked-disabled.png", rS)
	if err != nil {
		return nil, err
	}

	unchecked, err := loadGraphicImages("menu/checkbox-unchecked-idle.png", "menu/checkbox-unchecked-disabled.png", rS)
	if err != nil {
		return nil, err
	}

	greyed, err := loadGraphicImages("menu/checkbox-greyed-idle.png", "menu/checkbox-greyed-disabled.png", rS)
	if err != nil {
		return nil, err
	}

	return &checkboxResources{
		image: &widget.ButtonImage{
			Idle:     idle,
			Hover:    hover,
			Pressed:  hover,
			Disabled: disabled,
		},

		graphic: &widget.CheckboxGraphicImage{
			Checked:   checked,
			Unchecked: unchecked,
			Greyed:    greyed,
		},

		spacing: 5,
	}, nil
}

func newLabelResources(fonts *fonts) *labelResources {
	return &labelResources{
		text: &widget.LabelColor{
			Idle:     hexToColor(labelIdleColor),
			Disabled: hexToColor(labelDisabledColor),
		},

		face: fonts.face,
	}
}

func newComboButtonResources(fonts *fonts) (*comboButtonResources, error) {
	cH := centerHeightFromFontScale(fonts.scale)
	rS := resourceScaleFromFontScale(fonts.scale)
	idle, err := loadImageNineSlice("menu/combo-button-idle.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}

	hover, err := loadImageNineSlice("menu/combo-button-hover.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}

	pressed, err := loadImageNineSlice("menu/combo-button-pressed.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}

	disabled, err := loadImageNineSlice("menu/combo-button-disabled.png", 12, cH, rS)
	if err != nil {
		return nil, err
	}

	i := &widget.ButtonImage{
		Idle:     idle,
		Hover:    hover,
		Pressed:  pressed,
		Disabled: disabled,
	}

	// adjusted disabled combo box state to show as normal text color but not show selection arrow
	arrowDown, err := loadGraphicImages("menu/arrow-down-idle.png", "menu/arrow-down-hidden.png", rS)
	if err != nil {
		return nil, err
	}

	return &comboButtonResources{
		image: i,

		text: &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonIdleColor),
		},

		face:    fonts.face,
		graphic: arrowDown,

		padding: widget.Insets{
			Left:  30,
			Right: 30,
		},
	}, nil
}

func newListResources(fonts *fonts) (*listResources, error) {
	idle, _, err := resources.NewImageFromFile("menu/list-idle.png")
	if err != nil {
		return nil, err
	}

	disabled, _, err := resources.NewImageFromFile("menu/list-disabled.png")
	if err != nil {
		return nil, err
	}

	mask, _, err := resources.NewImageFromFile("menu/list-mask.png")
	if err != nil {
		return nil, err
	}

	trackIdle, _, err := resources.NewImageFromFile("menu/list-track-idle.png")
	if err != nil {
		return nil, err
	}

	trackDisabled, _, err := resources.NewImageFromFile("menu/list-track-disabled.png")
	if err != nil {
		return nil, err
	}

	handleIdle, _, err := resources.NewImageFromFile("menu/slider-handle-idle.png")
	if err != nil {
		return nil, err
	}

	handleHover, _, err := resources.NewImageFromFile("menu/slider-handle-hover.png")
	if err != nil {
		return nil, err
	}

	return &listResources{
		image: &widget.ScrollContainerImage{
			Idle:     image.NewNineSlice(idle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		},

		track: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
		},

		trackPadding: widget.Insets{
			Top:    5,
			Bottom: 24,
		},

		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
		},

		handleSize: 5,
		face:       fonts.face,

		entry: &widget.ListEntryColor{
			Unselected:         hexToColor(textIdleColor),
			DisabledUnselected: hexToColor(textDisabledColor),

			Selected:         hexToColor(textIdleColor),
			DisabledSelected: hexToColor(textDisabledColor),

			SelectedBackground:         hexToColor(listSelectedBackground),
			DisabledSelectedBackground: hexToColor(listDisabledSelectedBackground),

			FocusedBackground:         hexToColor(listFocusedBackground),
			SelectedFocusedBackground: hexToColor(listSelectedBackground),
		},

		entryPadding: widget.Insets{
			Left:   30,
			Right:  30,
			Top:    2,
			Bottom: 2,
		},
	}, nil
}

func newSliderResources() (*sliderResources, error) {
	idle, _, err := resources.NewImageFromFile("menu/slider-track-idle.png")
	if err != nil {
		return nil, err
	}

	disabled, _, err := resources.NewImageFromFile("menu/slider-track-disabled.png")
	if err != nil {
		return nil, err
	}

	handleIdle, _, err := resources.NewImageFromFile("menu/slider-handle-idle.png")
	if err != nil {
		return nil, err
	}

	handleHover, _, err := resources.NewImageFromFile("menu/slider-handle-hover.png")
	if err != nil {
		return nil, err
	}

	handleDisabled, _, err := resources.NewImageFromFile("menu/slider-handle-disabled.png")
	if err != nil {
		return nil, err
	}

	return &sliderResources{
		trackImage: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(idle, [3]int{0, 19, 0}, [3]int{6, 0, 0}),
			Hover:    image.NewNineSlice(idle, [3]int{0, 19, 0}, [3]int{6, 0, 0}),
			Disabled: image.NewNineSlice(disabled, [3]int{0, 19, 0}, [3]int{6, 0, 0}),
		},

		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleDisabled, 0, 5),
		},

		handleSize: 6,
	}, nil
}

func newPanelResources() (*panelResources, error) {
	i, err := loadImageNineSlice("menu/panel-idle.png", 10, 10, 1.0)
	if err != nil {
		return nil, err
	}
	f, err := loadImageNineSlice("menu/panel-filled.png", 10, 10, 1.0)
	if err != nil {
		return nil, err
	}
	t, err := loadImageNineSlice("menu/titlebar-idle.png", 10, 10, 1.0)
	if err != nil {
		return nil, err
	}
	return &panelResources{
		image:    i,
		filled:   f,
		titleBar: t,
		padding: widget.Insets{
			Left:   30,
			Right:  30,
			Top:    20,
			Bottom: 20,
		},
	}, nil
}

func newTabBookResources(fonts *fonts) (*tabBookResources, error) {

	return &tabBookResources{
		buttonFace: fonts.face,

		buttonText: &widget.ButtonTextColor{
			Idle:     hexToColor(buttonIdleColor),
			Disabled: hexToColor(buttonDisabledColor),
		},

		buttonPadding: widget.Insets{
			Left:  30,
			Right: 30,
		},
	}, nil
}

func newHeaderResources(fonts *fonts) (*headerResources, error) {
	bg, err := loadImageNineSlice("menu/header.png", 446, 9, 1.0)
	if err != nil {
		return nil, err
	}

	return &headerResources{
		background: bg,

		padding: widget.Insets{
			Left:   25,
			Right:  25,
			Top:    4,
			Bottom: 4,
		},

		face:  fonts.bigTitleFace,
		color: hexToColor(headerColor),
	}, nil
}

func newTextAreaResources(fonts *fonts) (*textAreaResources, error) {
	idle, _, err := resources.NewImageFromFile("menu/list-idle.png")
	if err != nil {
		return nil, err
	}

	disabled, _, err := resources.NewImageFromFile("menu/list-disabled.png")
	if err != nil {
		return nil, err
	}

	mask, _, err := resources.NewImageFromFile("menu/list-mask.png")
	if err != nil {
		return nil, err
	}

	trackIdle, _, err := resources.NewImageFromFile("menu/list-track-idle.png")
	if err != nil {
		return nil, err
	}

	trackDisabled, _, err := resources.NewImageFromFile("menu/list-track-disabled.png")
	if err != nil {
		return nil, err
	}

	handleIdle, _, err := resources.NewImageFromFile("menu/slider-handle-idle.png")
	if err != nil {
		return nil, err
	}

	handleHover, _, err := resources.NewImageFromFile("menu/slider-handle-hover.png")
	if err != nil {
		return nil, err
	}

	return &textAreaResources{
		image: &widget.ScrollContainerImage{
			Idle:     image.NewNineSlice(idle, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(disabled, [3]int{25, 12, 22}, [3]int{25, 12, 25}),
			Mask:     image.NewNineSlice(mask, [3]int{26, 10, 23}, [3]int{26, 10, 26}),
		},

		track: &widget.SliderTrackImage{
			Idle:     image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Hover:    image.NewNineSlice(trackIdle, [3]int{5, 0, 0}, [3]int{25, 12, 25}),
			Disabled: image.NewNineSlice(trackDisabled, [3]int{0, 5, 0}, [3]int{25, 12, 25}),
		},

		trackPadding: widget.Insets{
			Top:    5,
			Bottom: 24,
		},

		handle: &widget.ButtonImage{
			Idle:     image.NewNineSliceSimple(handleIdle, 0, 5),
			Hover:    image.NewNineSliceSimple(handleHover, 0, 5),
			Pressed:  image.NewNineSliceSimple(handleHover, 0, 5),
			Disabled: image.NewNineSliceSimple(handleIdle, 0, 5),
		},

		handleSize: 5,
		face:       fonts.face,

		entryPadding: widget.Insets{
			Left:   20,
			Right:  20,
			Top:    2,
			Bottom: 2,
		},
	}, nil
}

func newToolTipResources(fonts *fonts) (*toolTipResources, error) {
	bg, _, err := resources.NewImageFromFile("menu/tool-tip.png")
	if err != nil {
		return nil, err
	}

	return &toolTipResources{
		background: image.NewNineSlice(bg, [3]int{19, 6, 13}, [3]int{19, 5, 13}),

		padding: widget.Insets{
			Left:   15,
			Right:  15,
			Top:    10,
			Bottom: 10,
		},

		face:     fonts.toolTipFace,
		monoFace: fonts.toolTipMono,
		color:    hexToColor(toolTipColor),
	}, nil
}

func hexToColor(h string) color.Color {
	return hexToColorAlpha(h, 255)
}

func hexToColorAlpha(h string, alpha uint8) color.Color {
	u, err := strconv.ParseUint(h, 16, 0)
	if err != nil {
		panic(err)
	}

	return color.NRGBA{
		R: uint8(u & 0xff0000 >> 16),
		G: uint8(u & 0xff00 >> 8),
		B: uint8(u & 0xff),
		A: alpha,
	}
}
