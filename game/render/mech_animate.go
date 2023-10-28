package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/harbdog/raycaster-go/geom"
)

type MechAnimationIndex int

const (
	ANIMATE_STATIC MechAnimationIndex = -1
	ANIMATE_IDLE   MechAnimationIndex = 0
	ANIMATE_STRUT  MechAnimationIndex = 1
	// TODO: ANIMATE_SHUTDOWN, ANIMATE_JUMP?
	ANIMATE_DESTRUCT MechAnimationIndex = 2
	NUM_ANIMATIONS   MechAnimationIndex = 3
)

type MechSpriteAnimate struct {
	sheet            *ebiten.Image
	maxCols, maxRows int
	numColsAtRow     [NUM_ANIMATIONS]int
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
	uWidth, uHeight := int(float64(srcWidth)/float64(NUM_PARTS)), srcHeight

	uSize := uWidth
	if uHeight > uWidth {
		// adjust size to square it off as needed by the raycasting of sprites
		uSize = uHeight
	}

	// determine offsets for center/bottom within each frame
	centerX, bottomY := float64(uSize)/2-float64(uWidth)/2, float64(uSize-uHeight-1)

	// maxCols will be determined later based on how many frames needed by any single animation row
	maxRows, maxCols := int(NUM_ANIMATIONS), 1

	// separate out each limb part from source image
	srcParts := make([]*mechAnimatePart, int(NUM_PARTS))
	for c := 0; c < int(NUM_PARTS); c++ {
		x, y := c*uWidth, 0
		cellRect := image.Rect(x, y, x+uWidth, y+uHeight)
		cellImg := srcImage.SubImage(cellRect).(*ebiten.Image)
		srcParts[c] = &mechAnimatePart{image: cellImg, travelY: 0}
	}

	// calculate number of animations (rows) and frames for each animation (cols)
	numColsAtRow := [NUM_ANIMATIONS]int{}

	// idle animation: only arms and torso move, for now going with 4% pixel movement for both
	idlePxPerLimb := 0.02 * float64(uHeight)
	numColsAtRow[ANIMATE_IDLE] = 8 // 4x2 = up -> down -> down -> up (both arms only)
	if numColsAtRow[ANIMATE_IDLE] > maxCols {
		maxCols = numColsAtRow[ANIMATE_IDLE]
	}

	// strut animation: for now going with 2% for arms, 6% pixel movement for legs
	strutPxPerArm, strutPxPerLeg := 0.02*float64(uHeight), 0.06*float64(uHeight)
	numColsAtRow[ANIMATE_STRUT] = 16 // 4x4 = up -> down -> down -> up (starting with left arm, reverse right arm)
	if numColsAtRow[ANIMATE_STRUT] > maxCols {
		maxCols = numColsAtRow[ANIMATE_STRUT]
	}

	// destruction animation: for now arms and torso drop towards the ground 40% of the pixel height
	destructPxPerLimb := 0.4 * float64(uHeight)
	numColsAtRow[ANIMATE_DESTRUCT] = 16
	if numColsAtRow[ANIMATE_DESTRUCT] > maxCols {
		maxCols = numColsAtRow[ANIMATE_DESTRUCT]
	}

	mechSheet := ebiten.NewImage(maxCols*uSize, maxRows*uSize)

	m := &MechSpriteAnimate{
		sheet:        mechSheet,
		maxCols:      maxCols,
		maxRows:      maxRows,
		numColsAtRow: numColsAtRow,
	}

	// draw idle animation
	m.drawMechIdle(uSize, centerX, bottomY, idlePxPerLimb, srcParts)

	// draw strut animation
	m.drawMechStrut(uSize, centerX, bottomY, strutPxPerArm, strutPxPerLeg, srcParts)

	// draw destruction animation
	m.drawMechDestruction(uSize, centerX, bottomY, destructPxPerLimb, srcParts)

	return m
}

// drawMechIdle draws onto the sheet the idle animation in its assigned row in the sheet
func (m *MechSpriteAnimate) drawMechIdle(uSize int, adjustX, adjustY, pxPerLimb float64, parts []*mechAnimatePart) {
	row, col := int(ANIMATE_IDLE), 0

	resetMechAnimationParts(parts)
	ct := parts[PART_CT]
	la := parts[PART_LA]
	ra := parts[PART_RA]
	ll := parts[PART_LL]
	rl := parts[PART_RL]

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
	row, col := int(ANIMATE_STRUT), 0

	resetMechAnimationParts(parts)
	ct := parts[PART_CT]
	la := parts[PART_LA]
	ra := parts[PART_RA]
	ll := parts[PART_LL]
	rl := parts[PART_RL]

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

// drawMechDestruction draws onto the sheet the destruct animation in its assigned row in the sheet
func (m *MechSpriteAnimate) drawMechDestruction(uSize int, adjustX, adjustY, pxPerLimb float64, parts []*mechAnimatePart) {
	row, col := int(ANIMATE_DESTRUCT), 0

	resetMechAnimationParts(parts)
	ct := parts[PART_CT]
	la := parts[PART_LA]
	ra := parts[PART_RA]
	ll := parts[PART_LL]
	rl := parts[PART_RL]

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

	op_ct := &ebiten.DrawImageOptions{}
	op_ct.GeoM.Translate(offX, offY+offCT)
	m.sheet.DrawImage(ct, op_ct)

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
