package oth

import (
	"encoding/hex"
	"fmt"
)

func AbbrHexStr(b []byte, l int) string {
	if len(b) == 0 {
		return "-"
	}
	h := hex.EncodeToString(b)
	abrv := h[:Min(len(h), l)]
	if len(h) > l {
		abrv = fmt.Sprintf("%s...", abrv)
	}
	return abrv
}

func CheckPanic(e error) {
	if e != nil {
		panic(e)
	}
}
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
