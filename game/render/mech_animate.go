package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
)

type MechAnimationIndex int

const (
	MECH_ANIMATE_IDLE MechAnimationIndex = iota
	MECH_ANIMATE_STRUT
	// TODO: MECH_ANIMATE_JUMP
	MECH_ANIMATE_SHUTDOWN
	MECH_ANIMATE_DESTRUCT
	NUM_MECH_ANIMATIONS
	MECH_ANIMATE_STATIC MechAnimationIndex = -1
)

type MechSpriteAnimate struct {
	sheet            *ebiten.Image
	maxCols, maxRows int
	config           [NUM_MECH_ANIMATIONS]*mechAnimateConfig
}

type mechAnimateConfig struct {
	numCols       int
	animationRate int
	maxLoops      int
}

type mechAnimatePart struct {
	image   *ebiten.Image
	travelY float64
}

// NewMechAnimationSheetFromImage creates a new image sheet with generated image frames for mech sprite animation
func NewMechAnimationSheetFromImage(srcImage *ebiten.Image) *MechSpriteAnimate {
	// all mech sprite sheets have 6 columns of images in the sheet:
	// [full, torso, left arm, right arm, left leg, right leg]
	srcWidth, srcHeight := srcImage.Bounds().Dx(), srcImage.Bounds().Dy()
	uWidth, uHeight := int(float64(srcWidth)/float64(NUM_MECH_PARTS)), srcHeight

	uSize := uWidth
	if uHeight > uWidth {
		// adjust size to square it off as needed by the raycasting of sprites
		uSize = uHeight
	}

	// determine offsets for center/bottom within each frame
	centerX, bottomY := float64(uSize)/2-float64(uWidth)/2, float64(uSize-uHeight-1)

	// maxCols will be determined later based on how many frames needed by any single animation row
	maxRows, maxCols := int(NUM_MECH_ANIMATIONS), 1

	// separate out each limb part from source image
	srcParts := make([]*mechAnimatePart, int(NUM_MECH_PARTS))
	for c := 0; c < int(NUM_MECH_PARTS); c++ {
		x, y := c*uWidth, 0
		cellRect := image.Rect(x, y, x+uWidth, y+uHeight)
		cellImg := srcImage.SubImage(cellRect).(*ebiten.Image)
		srcParts[c] = &mechAnimatePart{image: cellImg, travelY: 0}
	}

	// calculate number of animations (rows) and frames for each animation (cols)
	animationCfg := [NUM_MECH_ANIMATIONS]*mechAnimateConfig{}

	// idle animation: only arms and torso move, for now going with 2% pixel movement for both
	idlePxPerLimb := 0.02 * float64(uHeight)
	animationCfg[MECH_ANIMATE_IDLE] = &mechAnimateConfig{
		numCols:       8, // 4x2 = up -> down -> down -> up (both arms only)
		animationRate: 7,
		maxLoops:      0,
	}
	if animationCfg[MECH_ANIMATE_IDLE].numCols > maxCols {
		maxCols = animationCfg[MECH_ANIMATE_IDLE].numCols
	}

	// strut animation: for now going with 2% for arms, 6% pixel movement for legs
	strutPxPerArm, strutPxPerLeg := 0.02*float64(uHeight), 0.06*float64(uHeight)
	animationCfg[MECH_ANIMATE_STRUT] = &mechAnimateConfig{
		numCols:       16, // 4x4 = up -> down -> down -> up (starting with left arm, reverse right arm)
		animationRate: 3,
		maxLoops:      0,
	}
	if animationCfg[MECH_ANIMATE_STRUT].numCols > maxCols {
		maxCols = animationCfg[MECH_ANIMATE_STRUT].numCols
	}

	// shut down animation: torso drops 8% pixel height, followed by arms dropping 12% pixel height
	downPxTorso, downPxPerArm := 0.08*float64(uHeight), 0.12*float64(uHeight)
	animationCfg[MECH_ANIMATE_SHUTDOWN] = &mechAnimateConfig{
		numCols:       12,
		animationRate: 8,
		maxLoops:      1,
	}
	if animationCfg[MECH_ANIMATE_SHUTDOWN].numCols > maxCols {
		maxCols = animationCfg[MECH_ANIMATE_SHUTDOWN].numCols
	}

	// destruction animation: for now arms and torso drop towards the ground 40% of the pixel height
	// TODO: make bigger mechs having longer destruction animations (more frames)
	destructPxPerLimb := 0.4 * float64(uHeight)
	animationCfg[MECH_ANIMATE_DESTRUCT] = &mechAnimateConfig{
		numCols:       16,
		animationRate: 2,
		maxLoops:      1,
	}
	if animationCfg[MECH_ANIMATE_DESTRUCT].numCols > maxCols {
		maxCols = animationCfg[MECH_ANIMATE_DESTRUCT].numCols
	}

	mechSheet := ebiten.NewImage(maxCols*uSize, maxRows*uSize)

	m := &MechSpriteAnimate{
		sheet:   mechSheet,
		maxCols: maxCols,
		maxRows: maxRows,
		config:  animationCfg,
	}

	// draw idle animation
	m.drawMechIdle(uSize, centerX, bottomY, idlePxPerLimb, srcParts)

	// draw strut animation
	m.drawMechStrut(uSize, centerX, bottomY, strutPxPerArm, strutPxPerLeg, srcParts)

	// draw shutdown animation
	m.drawMechShutdown(uSize, centerX, bottomY, downPxTorso, downPxPerArm, srcParts)

	// draw destruction animation
	m.drawMechDestruction(uSize, centerX, bottomY, destructPxPerLimb, srcParts)

	return m
}

// drawMechIdle draws onto the sheet the idle animation in its assigned row in the sheet
func (m *MechSpriteAnimate) drawMechIdle(uSize int, adjustX, adjustY, pxPerLimb float64, parts []*mechAnimatePart) {
	row, col := int(MECH_ANIMATE_IDLE), 0

	resetMechAnimationParts(parts)
	ct := parts[MECH_PART_CT]
	la := parts[MECH_PART_LA]
	ra := parts[MECH_PART_RA]
	ll := parts[MECH_PART_LL]
	rl := parts[MECH_PART_RL]

	// first frame of idle animation is static image
	m.drawMechAnimationParts(row, col, 1, uSize, adjustX, adjustY, ct, 0, la, 0, 0, ra, 0, 0, ll, 0, 0, rl, 0, 0)
	col++

	// 2x arms up
	m.drawMechAnimationParts(
		row, col, 2, uSize, adjustX, adjustY, ct, 0, la, -pxPerLimb, 0, ra, -pxPerLimb, 0, ll, 0, 0, rl, 0, 0,
	)
	col += 2

	// 2x arms down
	m.drawMechAnimationParts(
		row, col, 2, uSize, adjustX, adjustY, ct, 0, la, pxPerLimb, 0, ra, pxPerLimb, 0, ll, 0, 0, rl, 0, 0,
	)
	col += 2

	// 2x arms down + 2x ct down
	m.drawMechAnimationParts(
		row, col, 2, uSize, adjustX, adjustY, ct, pxPerLimb, la, pxPerLimb, 0, ra, pxPerLimb, 0, ll, 0, 0, rl, 0, 0,
	)
	col += 2

	// 1x arms and ct back up again
	m.drawMechAnimationParts(
		row, col, 1, uSize, adjustX, adjustY, ct, -pxPerLimb/2, la, -pxPerLimb/2, 0, ra, -pxPerLimb/2, 0, ll, 0, 0, rl, 0, 0,
	)
}

// drawMechStrut draws onto the sheet the strut animation in its assigned row in the sheet
func (m *MechSpriteAnimate) drawMechStrut(uSize int, adjustX, adjustY, pxPerArm, pxPerLeg float64, parts []*mechAnimatePart) {
	row, col := int(MECH_ANIMATE_STRUT), 0

	resetMechAnimationParts(parts)
	ct := parts[MECH_PART_CT]
	la := parts[MECH_PART_LA]
	ra := parts[MECH_PART_RA]
	ll := parts[MECH_PART_LL]
	rl := parts[MECH_PART_RL]

	// ct needs to also move up as a leg moves up, half as much
	pxPerTorso := pxPerLeg / 2

	// 4x rl/la up + ra down
	m.drawMechAnimationParts(
		row, col, 4, uSize, adjustX, adjustY, ct, -pxPerTorso, la, -pxPerArm, 0, ra, pxPerArm, 0, ll, 0, 0, rl, -pxPerLeg, 0,
	)
	col += 4

	// 4x rl/la down + ra up
	m.drawMechAnimationParts(
		row, col, 4, uSize, adjustX, adjustY, ct, pxPerTorso, la, pxPerArm, 0, ra, -pxPerArm, 0, ll, 0, 0, rl, pxPerLeg, 0,
	)
	col += 4

	// 4x ll/ra up + la down
	m.drawMechAnimationParts(
		row, col, 4, uSize, adjustX, adjustY, ct, -pxPerTorso, la, pxPerArm, 0, ra, -pxPerArm, 0, ll, -pxPerLeg, 0, rl, 0, 0,
	)
	col += 4

	// 4x ll/ra down + la up
	m.drawMechAnimationParts(
		row, col, 4, uSize, adjustX, adjustY, ct, pxPerTorso, la, -pxPerArm, 0, ra, pxPerArm, 0, ll, pxPerLeg, 0, rl, 0, 0,
	)
}

// drawMechShutdown draws onto the sheet the idle animation in its assigned row in the sheet
func (m *MechSpriteAnimate) drawMechShutdown(uSize int, adjustX, adjustY, downPxTorso, downPxPerArm float64, parts []*mechAnimatePart) {
	row, col := int(MECH_ANIMATE_SHUTDOWN), 0

	resetMechAnimationParts(parts)
	ct := parts[MECH_PART_CT]
	la := parts[MECH_PART_LA]
	ra := parts[MECH_PART_RA]
	ll := parts[MECH_PART_LL]
	rl := parts[MECH_PART_RL]

	// 3x torso down
	m.drawMechAnimationParts(
		row, col, 3, uSize, adjustX, adjustY, ct, downPxTorso/2, la, 0, 0, ra, 0, 0, ll, 0, 0, rl, 0, 0,
	)
	col += 3

	// 6x torso down with arms down
	m.drawMechAnimationParts(
		row, col, 6, uSize, adjustX, adjustY, ct, downPxTorso, la, downPxPerArm, 0, ra, downPxPerArm, 0, ll, 0, 0, rl, 0, 0,
	)
	col += 6

	// 3x arms down
	m.drawMechAnimationParts(
		row, col, 3, uSize, adjustX, adjustY, ct, 0, la, downPxPerArm/2, 0, ra, downPxPerArm/2, 0, ll, 0, 0, rl, 0, 0,
	)
	col += 3
}

// drawMechDestruction draws onto the sheet the destruct animation in its assigned row in the sheet
func (m *MechSpriteAnimate) drawMechDestruction(uSize int, adjustX, adjustY, pxPerLimb float64, parts []*mechAnimatePart) {
	row, col := int(MECH_ANIMATE_DESTRUCT), 0

	resetMechAnimationParts(parts)
	ct := parts[MECH_PART_CT]
	la := parts[MECH_PART_LA]
	ra := parts[MECH_PART_RA]
	ll := parts[MECH_PART_LL]
	rl := parts[MECH_PART_RL]

	// arms and ct drop all the way down with limb rotation for arms and legs falling off
	rotPerLeft, rotPerRight := -geom.HalfPi, geom.HalfPi
	m.drawMechAnimationParts(
		row, col, 16, uSize, adjustX, adjustY, ct, pxPerLimb, la, pxPerLimb/2, rotPerLeft, ra, pxPerLimb/2, rotPerRight, ll, pxPerLimb/2, rotPerLeft, rl, pxPerLimb/2, rotPerRight,
	)
}

func resetMechAnimationParts(parts []*mechAnimatePart) {
	for _, part := range parts {
		part.travelY = 0
	}
}

// drawMechAnimationParts draws onto the sheet each mech part with total pixel travel and rotation (radians)
// over a number of given frames starting at the given column within the given row in the sheet of frames
func (m *MechSpriteAnimate) drawMechAnimationParts(
	row, col, frames, uSize int, adjustX, adjustY float64, ct *mechAnimatePart, pxCT float64,
	la *mechAnimatePart, pxLA, rotLA float64, ra *mechAnimatePart, pxRA, rotRA float64,
	ll *mechAnimatePart, pxLL, rotLL float64, rl *mechAnimatePart, pxRL, rotRL float64,
) {
	offsetY := float64(row*uSize) + adjustY

	// use previously tracked offsets in parts as starting point
	pxPerCT := ct.travelY
	pxPerLA := la.travelY
	pxPerRA := ra.travelY
	pxPerLL := ll.travelY
	pxPerRL := rl.travelY

	// rotPerCT not necessary, limb rotation only for destruct animation
	var rotPerLA, rotPerRA, rotPerLL, rotPerRL float64

	for c := col; c < col+frames; c++ {
		offsetX := float64(c*uSize) + adjustX
		pxPerCT += pxCT / float64(frames)
		pxPerLA += pxLA / float64(frames)
		pxPerRA += pxRA / float64(frames)
		pxPerLL += pxLL / float64(frames)
		pxPerRL += pxRL / float64(frames)

		rotPerLA += rotLA / float64(frames)
		rotPerRA += rotRA / float64(frames)
		rotPerLL += rotLL / float64(frames)
		rotPerRL += rotRL / float64(frames)

		m.drawMechAnimFrame(
			offsetX, offsetY, ct.image, pxPerCT,
			la.image, pxPerLA, rotPerLA,
			ra.image, pxPerRA, rotPerRA,
			ll.image, pxPerLL, rotPerLL,
			rl.image, pxPerRL, rotPerRL,
		)
	}

	// keep track of offsets in parts for next animation
	ct.travelY += pxCT
	la.travelY += pxLA
	ra.travelY += pxRA
	ll.travelY += pxLL
	rl.travelY += pxRL
}

// drawMechAnimFrame draws onto the sheet each mech part each with given offet for the frame (offX, offY),
// and individual offsets and rotations specific for each part
func (m *MechSpriteAnimate) drawMechAnimFrame(
	offX, offY float64, ct *ebiten.Image, offCT float64,
	la *ebiten.Image, offLA, rotLA float64,
	ra *ebiten.Image, offRA, rotRA float64,
	ll *ebiten.Image, offLL, rotLL float64,
	rl *ebiten.Image, offRL, rotRL float64,
) {
	w, h := ct.Bounds().Dx(), ct.Bounds().Dy()

	op_ll := &ebiten.DrawImageOptions{}
	if rotLL != 0 {
		op_ll.GeoM.Translate(-float64(w)/2, -float64(h/2))
		op_ll.GeoM.Rotate(rotLL)
		op_ll.GeoM.Translate(float64(w)/2, float64(h/2))
	}
	op_ll.GeoM.Translate(offX, offY+offLL)
	m.sheet.DrawImage(ll, op_ll)

	op_rl := &ebiten.DrawImageOptions{}
	if rotRL != 0 {
		op_rl.GeoM.Translate(-float64(w)/2, -float64(h/2))
		op_rl.GeoM.Rotate(rotRL)
		op_rl.GeoM.Translate(float64(w)/2, float64(h/2))
	}
	op_rl.GeoM.Translate(offX, offY+offRL)
	m.sheet.DrawImage(rl, op_rl)

	op_ct := &ebiten.DrawImageOptions{}
	op_ct.GeoM.Translate(offX, offY+offCT)
	m.sheet.DrawImage(ct, op_ct)

	op_la := &ebiten.DrawImageOptions{}
	if rotLA != 0 {
		op_la.GeoM.Translate(-float64(w)/2, -float64(h/2))
		op_la.GeoM.Rotate(rotLA)
		op_la.GeoM.Translate(float64(w)/2, float64(h/2))
	}
	op_la.GeoM.Translate(offX, offY+offLA)
	m.sheet.DrawImage(la, op_la)

	op_ra := &ebiten.DrawImageOptions{}
	if rotRA != 0 {
		op_ra.GeoM.Translate(-float64(w)/2, -float64(h/2))
		op_ra.GeoM.Rotate(rotRA)
		op_ra.GeoM.Translate(float64(w)/2, float64(h/2))
	}
	op_ra.GeoM.Translate(offX, offY+offRA)
	m.sheet.DrawImage(ra, op_ra)
}
