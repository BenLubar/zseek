package zseek

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func testSetup(t *testing.T) (z *ZSeek, cleanup func()) {
	f, err := ioutil.TempFile(os.TempDir(), "zseek")
	if err != nil {
		t.Skip("temp file creation failed: ", err)
		t.SkipNow()
	}

	z, err = New(f)
	if err != nil {
		t.Skip("New failed: ", err)
		t.SkipNow()
	}

	cleanup = func() {
		err := z.Close()
		if err != nil {
			t.Error("closing ZSeek: ", err)
		}
		name := f.Name()
		err = f.Close()
		if err != nil {
			t.Error("closing file: ", err)
		}
		err = os.Remove(name)
		if err != nil {
			t.Error("removing file: ", err)
		}
	}
	return
}

func TestEmpty(t *testing.T) {
	z, cleanup := testSetup(t)
	defer cleanup()

	b := make([]byte, 10)

	n, err := z.Read(b)
	if n != 0 {
		t.Error("expected 0 bytes, but got ", n, " bytes")
	}
	if err != io.EOF {
		t.Error("expected io.EOF, but got ", err)
	}

	n, err = z.Read(b[:0])
	if n != 0 {
		t.Error("expected 0 bytes, but got ", n, " bytes")
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	o, err := z.Seek(0, SeekStart)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	o, err = z.Seek(0, SeekCur)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	o, err = z.Seek(0, SeekEnd)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	o, err = z.Seek(-1, SeekStart)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != ErrInvalidSeek {
		t.Error("expected ErrInvalidSeek, but got ", err)
	}

	o, err = z.Seek(-1, SeekCur)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != ErrInvalidSeek {
		t.Error("expected ErrInvalidSeek, but got ", err)
	}

	o, err = z.Seek(-1, SeekEnd)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != ErrInvalidSeek {
		t.Error("expected ErrInvalidSeek, but got ", err)
	}

	o, err = z.Seek(1, SeekStart)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != ErrInvalidSeek {
		t.Error("expected ErrInvalidSeek, but got ", err)
	}

	o, err = z.Seek(1, SeekCur)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != ErrInvalidSeek {
		t.Error("expected ErrInvalidSeek, but got ", err)
	}

	o, err = z.Seek(1, SeekEnd)
	if o != 0 {
		t.Error("expected offset 0, but got offset ", o)
	}
	if err != ErrInvalidSeek {
		t.Error("expected ErrInvalidSeek, but got ", err)
	}
}

func TestWrite(t *testing.T) {
	z, cleanup := testSetup(t)
	defer cleanup()

	b := make([]byte, 1024*1024)

	n, err := z.Write(b)
	if n != 1024*1024 {
		t.Error("expected ", 1024*1024, " bytes, but got ", n, " bytes")
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	o, err := z.Seek(0, SeekCur)
	if o != 1024*1024 {
		t.Error("expected offset ", 1024*1024, ", but got offset ", o)
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	o, err = z.Seek(-1, SeekCur)
	if o != 1024*1024-1 {
		t.Error("expected offset ", 1024*1024-1, ", but got offset ", o)
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	n, err = z.Read(b)
	if n != 1 {
		t.Error("expected ", 1, " bytes, but got ", n, " bytes")
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	n, err = z.Read(b)
	if n != 0 {
		t.Error("expected ", 1, " bytes, but got ", n, " bytes")
	}
	if err != io.EOF {
		t.Error("expected io.EOF, but got ", err)
	}

	o, err = z.Seek(0, SeekEnd)
	if o != 1024*1024 {
		t.Error("expected offset ", 1024*1024, ", but got offset ", o)
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	err = z.Flush()
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	z2, err := New(z.f)
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	o, err = z2.Seek(0, SeekEnd)
	if o != 1024*1024 {
		t.Error("expected offset ", 1024*1024, ", but got offset ", o)
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	z2, err = New(z.f)
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	n, err = io.ReadFull(z2, b)
	if n != 1024*1024 {
		t.Error("expected ", 1024*1024, " bytes, but got ", n, " bytes")
	}
	if err != nil {
		t.Error("expected nil, but got ", err)
	}

	n, err = io.ReadFull(z2, b)
	if n != 0 {
		t.Error("expected ", 0, " bytes, but got ", n, " bytes")
	}
	if err != io.EOF {
		t.Error("expected io.EOF, but got ", err)
	}
}
