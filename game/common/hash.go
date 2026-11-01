package common

import (
	"crypto/md5"
	"fmt"
	"hash/fnv"
)

func GetFNVHash(str string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}

func GetMD5Sum(str string) string {
	data := []byte(str)
	md5str := fmt.Sprintf("%x", md5.Sum(data))
	return md5str
}
