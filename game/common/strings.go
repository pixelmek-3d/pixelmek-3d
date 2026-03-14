package common

import (
	"fmt"
	"strings"
)

// FloatDisplayString returns a displayable string value of a float with any trailing zeros/periods removed.
func FloatDisplayString(f float64) string {
	dispStr := fmt.Sprintf("%0.2f", f)
	dispStr = strings.TrimRight(dispStr, "0")
	dispStr = strings.TrimRight(dispStr, ".")
	return dispStr
}
