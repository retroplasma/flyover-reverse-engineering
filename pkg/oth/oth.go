package oth

import (
	"encoding/hex"
	"fmt"
)

// AbbrHexStr produces an abbreviated hex string from b and adds "..." if it's longer than l
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

// AbbrStr produces an abbreviated string from s and adds "..." if it's longer than l
func AbbrStr(s string, l int) string {
	if len(s) == 0 {
		return "-"
	}
	abrv := s[:Min(len(s), l)]
	if len(s) > l {
		abrv = fmt.Sprintf("%s...", abrv)
	}
	return abrv
}

// CheckPanic panics if e is not null
func CheckPanic(e error) {
	if e != nil {
		panic(e)
	}
}

// Min returns the minimum of two integers
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Max returns the maximum of two integers
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}
