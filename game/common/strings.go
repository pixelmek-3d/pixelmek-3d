package common

import (
	"fmt"
	"strings"
	"time"
)

// FloatDisplayString returns a displayable string value of a float with any trailing zeros/periods removed.
func FloatDisplayString(f float64) string {
	dispStr := fmt.Sprintf("%0.2f", f)
	dispStr = strings.TrimRight(dispStr, "0")
	dispStr = strings.TrimRight(dispStr, ".")
	return dispStr
}

// DurationDisplayString returns a displayable string value of a duration in hours:minutes:seconds (e.g. `32:09:14“)
func DurationDisplayString(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	str := fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	return str
}
