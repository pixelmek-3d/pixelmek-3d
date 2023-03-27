package game

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

type LogFormat struct {
	TimestampFormat string
}

func (f *LogFormat) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	b.WriteString(entry.Time.Format(f.TimestampFormat))

	b.WriteString(" [" + strings.ToUpper(entry.Level.String()) + "] ")

	if entry.Message != "" {
		b.WriteString(entry.Message)
	}

	if len(entry.Data) > 0 {
		b.WriteString(" || ")
	}
	for key, value := range entry.Data {
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteByte('{')
		fmt.Fprint(b, value)
		b.WriteString("}, ")
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}
