package render

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type MechAnimationIndex int

const (
	ANIMATE_STATIC MechAnimationIndex = -1
	ANIMATE_IDLE   MechAnimationIndex = 0
	ANIMATE_STRUT  MechAnimationIndex = 1
	// TODO: ANIMATE_SHUTDOWN, ANIMATE_JUMP?
	NUM_ANIMATIONS MechAnimationIndex = 2
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
	srcWidth, srcHeight := srcImage.Size()
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
	m.drawMechAnimationParts(row, col, 1, uSize, adjustX, adjustY, ct, 0, la, 0, ra, 0, ll, 0, rl, 0)
	col++

	// 2x arms up
	m.drawMechAnimationParts(row, col, 2, uSize, adjustX, adjustY, ct, 0, la, -pxPerLimb, ra, -pxPerLimb, ll, 0, rl, 0)
	col += 2

	// 2x arms down
	m.drawMechAnimationParts(row, col, 2, uSize, adjustX, adjustY, ct, 0, la, pxPerLimb, ra, pxPerLimb, ll, 0, rl, 0)
	col += 2

	// 2x arms down + 2x ct down
	m.drawMechAnimationParts(
		row, col, 2, uSize, adjustX, adjustY, ct, pxPerLimb, la, pxPerLimb, ra, pxPerLimb, ll, 0, rl, 0,
	)
	col += 2

	// 1x arms and ct back up again
	m.drawMechAnimationParts(
		row, col, 1, uSize, adjustX, adjustY, ct, -pxPerLimb/2, la, -pxPerLimb/2, ra, -pxPerLimb/2, ll, 0, rl, 0,
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
		row, col, 4, uSize, adjustX, adjustY, ct, -pxPerTorso, la, -pxPerArm, ra, pxPerArm, ll, 0, rl, -pxPerLeg,
	)
	col += 4

	// 4x rl/la down + ra up
	m.drawMechAnimationParts(
		row, col, 4, uSize, adjustX, adjustY, ct, pxPerTorso, la, pxPerArm, ra, -pxPerArm, ll, 0, rl, pxPerLeg,
	)
	col += 4

	// 4x ll/ra up + la down
	m.drawMechAnimationParts(
		row, col, 4, uSize, adjustX, adjustY, ct, -pxPerTorso, la, pxPerArm, ra, -pxPerArm, ll, -pxPerLeg, rl, 0,
	)
	col += 4

	// 4x ll/ra down + la up
	m.drawMechAnimationParts(
		row, col, 4, uSize, adjustX, adjustY, ct, pxPerTorso, la, -pxPerArm, ra, pxPerArm, ll, pxPerLeg, rl, 0,
	)
}

func resetMechAnimationParts(parts []*mechAnimatePart) {
	for _, part := range parts {
		part.travelY = 0
	}
}

// drawMechAnimationParts draws onto the sheet each mech part with total pixel travel over a number of given frames
// starting at the given column within the given row in the sheet of frames
func (m *MechSpriteAnimate) drawMechAnimationParts(
	row, col, frames, uSize int, adjustX, adjustY float64, ct *mechAnimatePart, pxCT float64,
	la *mechAnimatePart, pxLA float64, ra *mechAnimatePart, pxRA float64,
	ll *mechAnimatePart, pxLL float64, rl *mechAnimatePart, pxRL float64,
) {
	offsetY := float64(row*uSize) + adjustY

	// use previously tracked offsets in parts as starting point
	pxPerCT := ct.travelY
	pxPerLA := la.travelY
	pxPerRA := ra.travelY
	pxPerLL := ll.travelY
	pxPerRL := rl.travelY

	for c := col; c < col+frames; c++ {
		offsetX := float64(c*uSize) + adjustX
		pxPerCT += pxCT / float64(frames)
		pxPerLA += pxLA / float64(frames)
		pxPerRA += pxRA / float64(frames)
		pxPerLL += pxLL / float64(frames)
		pxPerRL += pxRL / float64(frames)

		m.drawMechAnimFrame(
			offsetX, offsetY, ct.image, pxPerCT, la.image, pxPerLA,
			ra.image, pxPerRA, ll.image, pxPerLL, rl.image, pxPerRL,
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
// and individual offsets specific for each part
func (m *MechSpriteAnimate) drawMechAnimFrame(
	offX, offY float64, ct *ebiten.Image, offCT float64, la *ebiten.Image, offLA float64,
	ra *ebiten.Image, offRA float64, ll *ebiten.Image, offLL float64, rl *ebiten.Image, offRL float64,
) {
	offset := ebiten.GeoM{}
	offset.Translate(offX, offY)

	op_ct := &ebiten.DrawImageOptions{GeoM: offset}
	op_ct.GeoM.Translate(0, offCT)
	m.sheet.DrawImage(ct, op_ct)

	op_ll := &ebiten.DrawImageOptions{GeoM: offset}
	op_ll.GeoM.Translate(0, offLL)
	m.sheet.DrawImage(ll, op_ll)

	op_rl := &ebiten.DrawImageOptions{GeoM: offset}
	op_rl.GeoM.Translate(0, offRL)
	m.sheet.DrawImage(rl, op_rl)

	op_la := &ebiten.DrawImageOptions{GeoM: offset}
	op_la.GeoM.Translate(0, offLA)
	m.sheet.DrawImage(la, op_la)

	op_ra := &ebiten.DrawImageOptions{GeoM: offset}
	op_ra.GeoM.Translate(0, offRA)
	m.sheet.DrawImage(ra, op_ra)
}
