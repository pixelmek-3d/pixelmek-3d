package common

import (
	"crypto/md5"
	"fmt"
)

func GetMD5Sum(str string) string {
	data := []byte(str)
	md5str := fmt.Sprintf("%x", md5.Sum(data))
	return md5str
}
