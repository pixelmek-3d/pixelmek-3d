package transitions

type TransitionDirection int

const (
	TransitionIn TransitionDirection = iota
	TransitionHold
	TransitionOut
)

type TransitionOptions struct {
	CurrentDirection TransitionDirection
	InDuration       float32
	HoldDuration     float32
	OutDuration      float32
}

func (t *TransitionOptions) Duration() float32 {
	switch t.CurrentDirection {
	case TransitionIn:
		return t.InDuration
	case TransitionHold:
		return t.HoldDuration
	case TransitionOut:
		return t.OutDuration
	}

	return 0
}
