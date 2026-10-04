//go:build windows

package clr

import (
	"bytes"
	"runtime"
	"testing"
	"unsafe"
)

func TestUTF16LE(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{"empty", "", nil},
		{"ASCII", "Go", []byte{0x47, 0x00, 0x6f, 0x00}},
		{"BMP", "é中", []byte{0xe9, 0x00, 0x2d, 0x4e}},
		{"surrogate pair", "😀", []byte{0x3d, 0xd8, 0x00, 0xde}},
		{"embedded NUL", "a\x00b", []byte{0x61, 0x00, 0x00, 0x00, 0x62, 0x00}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utf16Le(tt.input); !bytes.Equal(got, tt.want) {
				t.Errorf("utf16Le(%q) = % x, want % x", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpectsParams(t *testing.T) {
	tests := []struct {
		signature string
		want      bool
	}{
		{"Void Main()", false},
		{"Void Main(System.String[])", true},
		{"Int32 Main(System.String[])", true},
	}
	for _, tt := range tests {
		t.Run(tt.signature, func(t *testing.T) {
			if got := expectsParams(tt.signature); got != tt.want {
				t.Errorf("expectsParams(%q) = %t, want %t", tt.signature, got, tt.want)
			}
		})
	}
}

func TestCheckOK(t *testing.T) {
	if err := checkOK(S_OK, "test caller"); err != nil {
		t.Fatalf("checkOK(S_OK) = %v, want nil", err)
	}
	for _, tt := range []struct {
		hr   uintptr
		want string
	}{
		{S_FALSE, "test caller returned 0x00000001"},
		{uintptr(E_POINTER), "test caller returned 0x80004003"},
	} {
		if err := checkOK(tt.hr, "test caller"); err == nil || err.Error() != tt.want {
			t.Errorf("checkOK(0x%x) = %v, want %q", tt.hr, err, tt.want)
		}
	}
}

func TestReadUnicodeStr(t *testing.T) {
	tests := []struct {
		name  string
		input []uint16
		want  string
	}{
		{"empty", []uint16{0, 'X'}, ""},
		{"ASCII and trailing data", []uint16{'G', 'o', 0, 'X'}, "Go"},
		{"BMP", []uint16{0x00e9, 0x4e2d, 0}, "é中"},
		{"surrogate pair", []uint16{0xd83d, 0xde00, 0, 'X'}, "😀"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReadUnicodeStr(unsafe.Pointer(&tt.input[0]))
			runtime.KeepAlive(tt.input)
			if got != tt.want {
				t.Errorf("ReadUnicodeStr(%x) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
