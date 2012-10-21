package vpk

import (
	"crypto/sha1"
	"io"
	"os"
	"reflect"
	"testing"
)

var emptyVPKHash = []byte{0xa3, 0x48, 0x36, 0xea, 0x5b, 0xb9, 0x9b, 0x17, 0x08, 0x1f, 0xbd, 0x4f, 0xb9, 0x62, 0xcf, 0x37, 0xbc, 0x63, 0xd8, 0x22}

func TestInvalid(t *testing.T) {
	const filename = "testdata/notreallya.vpk"

	in, err := os.Open(filename)
	if err != nil {
		t.Error(err)
		return
	}
	defer in.Close()

	_, err = ReadVPKFile(in)
	if err == nil {
		t.Error("ReadVPKFile on an invalid VPK file did not return an error")
	} else {
		if expected := "Invalid signature 0x73696854 (expected 0x55aa1234)"; expected != err.Error() {
			t.Error(err)
			t.Log("Expected:", expected)
		}
	}
}

func TestEmpty(t *testing.T) {
	const filename = "testdata/empty.vpk"

	in, err := os.Open(filename)
	if err != nil {
		t.Error(err)
		return
	}
	defer in.Close()

	f, err := ReadVPKFile(in)
	if err != nil {
		t.Error(err)
		return
	}

	for _, file := range f.ListFiles() {
		t.Errorf("Unexpected file in %s: %q", filename, file)
	}

	if info := f.GetFileInfo("some/file.txt"); info != nil {
		t.Errorf("Nonexistent file returned a FileInfo: %#v", *info)
	}
}

func TestSingleFile(t *testing.T) {
	const filename = "testdata/singlefile.vpk"

	in, err := os.Open(filename)
	if err != nil {
		t.Error(err)
		return
	}
	defer in.Close()

	f, err := ReadVPKFile(in)
	if err != nil {
		t.Error(err)
		return
	}

	found := 0
	for _, file := range f.ListFiles() {
		if file == "empty.vpk" {
			found++
		} else {
			t.Errorf("Unexpected file in %s: %q", filename, file)
		}
	}

	if found != 1 {
		t.Errorf("%s had %d files named empty.vpk, but %d were expected", filename, found, 1)
	}

	if info := f.GetFileInfo("some/file.txt"); info != nil {
		t.Errorf("Nonexistent file returned a FileInfo: %#v", *info)
	}

	if info := f.GetFileInfo("empty.vpk"); info != nil {
		r, err := f.GetReader(info, filename)
		if err != nil {
			t.Error(err)
			return
		}

		h := sha1.New()
		io.Copy(h, r)

		if hash := h.Sum(nil); !reflect.DeepEqual(hash, emptyVPKHash) {
			t.Logf("Expected: % x", emptyVPKHash)
			t.Logf("Found   : % x", hash)
			t.Error("empty.vpk hash did not match expected!")
		}

		r.Close()
	} else {
		t.Error("No FileInfo for empty.vpk")
	}
}
