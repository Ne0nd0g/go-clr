//go:build windows

package clr

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

func destroyTestSafeArray(t *testing.T, array *SafeArray) {
	t.Helper()
	t.Cleanup(func() {
		if err := SafeArrayDestroy(array); err != nil {
			t.Errorf("SafeArrayDestroy: %v", err)
		}
	})
}

func readTestSafeArray(t *testing.T, array *SafeArray, size int) []byte {
	t.Helper()
	data, err := SafeArrayAccessData(array)
	if err != nil {
		t.Fatalf("SafeArrayAccessData: %v", err)
	}
	// AccessData must be paired with UnaccessData, including on assertion failure.
	// https://learn.microsoft.com/en-us/windows/win32/api/oleauto/nf-oleauto-safearrayaccessdata
	defer func() {
		unaccess := windows.NewLazySystemDLL("oleaut32.dll").NewProc("SafeArrayUnaccessData")
		hr, _, _ := unaccess.Call(uintptr(unsafe.Pointer(array)))
		if hr != S_OK {
			t.Errorf("SafeArrayUnaccessData HRESULT = %#x", hr)
		}
	}()
	if data == nil {
		t.Fatal("SafeArrayAccessData returned nil data")
	}
	return bytes.Clone(unsafe.Slice((*byte)(unsafe.Pointer(data)), size))
}

func TestCreateSafeArray(t *testing.T) {
	for _, raw := range [][]byte{{0x00}, {0x00, 0x01, 0x7f, 0x80, 0xff}, []byte("CLR byte array")} {
		t.Run(fmt.Sprintf("%x", raw), func(t *testing.T) {
			original := bytes.Clone(raw)
			array, err := CreateSafeArray(raw)
			if err != nil {
				t.Fatalf("CreateSafeArray: %v", err)
			}
			if array == nil {
				t.Fatal("CreateSafeArray returned nil")
			}
			destroyTestSafeArray(t, array)
			vt, err := SafeArrayGetVartype(array)
			if err != nil || vt != VT_UI1 {
				t.Fatalf("SafeArrayGetVartype = %#x, %v; want VT_UI1", vt, err)
			}
			if array.cDims != 1 || array.cbElements != 1 {
				t.Fatalf("dimensions = %d, element size = %d; want 1, 1", array.cDims, array.cbElements)
			}
			lower, err := SafeArrayGetLBound(array, 1)
			if err != nil || lower != 0 {
				t.Fatalf("lower bound = %d, %v; want 0", lower, err)
			}
			upper, err := SafeArrayGetUBound(array, 1)
			if err != nil || upper != uint32(len(raw)-1) {
				t.Fatalf("upper bound = %d, %v; want %d", upper, err, len(raw)-1)
			}
			if got := readTestSafeArray(t, array, len(raw)); !bytes.Equal(got, raw) {
				t.Errorf("array data = %x; want %x", got, raw)
			}
			raw[0] ^= 0xff
			if got := readTestSafeArray(t, array, len(raw)); !bytes.Equal(got, original) {
				t.Errorf("array data after source mutation = %x; want %x", got, original)
			}
			if array.cLocks != 0 {
				t.Errorf("lock count after read = %d; want 0", array.cLocks)
			}
		})
	}
}

func TestSafeArrayPutElement(t *testing.T) {
	bounds := SafeArrayBound{cElements: 3, lLbound: 4}
	array, err := SafeArrayCreate(VT_UI1, 1, &bounds)
	if err != nil {
		t.Fatalf("SafeArrayCreate: %v", err)
	}
	if array == nil {
		t.Fatal("SafeArrayCreate returned nil")
	}
	destroyTestSafeArray(t, array)
	lower, err := SafeArrayGetLBound(array, 1)
	if err != nil || lower != 4 {
		t.Fatalf("lower bound = %d, %v; want 4", lower, err)
	}
	upper, err := SafeArrayGetUBound(array, 1)
	if err != nil || upper != 6 {
		t.Fatalf("upper bound = %d, %v; want 6", upper, err)
	}
	value := byte(0xab)
	for _, index := range []int32{3, 7} {
		if err := SafeArrayPutElement(array, index, unsafe.Pointer(&value)); err == nil {
			t.Fatalf("SafeArrayPutElement accepted out-of-bounds index %d", index)
		}
	}
	if err := SafeArrayPutElement(array, 5, unsafe.Pointer(&value)); err != nil {
		t.Fatalf("SafeArrayPutElement: %v", err)
	}
	if got := readTestSafeArray(t, array, 3); !bytes.Equal(got, []byte{0, value, 0}) {
		t.Errorf("array data = %x; want 00ab00", got)
	}
}
