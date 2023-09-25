package conv

import "unsafe"

// QuickS2B quickly convert the string type to the []byte type
func QuickS2B(str string) []byte {
	base := *(*[2]uintptr)(unsafe.Pointer(&str))
	return *(*[]byte)(unsafe.Pointer(&[3]uintptr{base[0], base[1], base[1]}))
}

// QuickB2S quickly convert []byte to string
func QuickB2S(bs []byte) string {
	base := (*[3]uintptr)(unsafe.Pointer(&bs))
	return *(*string)(unsafe.Pointer(&[2]uintptr{base[0], base[1]}))
}
