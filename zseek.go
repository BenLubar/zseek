package zseek

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"sort"
)

// Constants for Seek
const (
	SeekStart = 0
	SeekCur   = 1
	SeekEnd   = 2
)

// Constants for NewLevel and NewBuffer
const (
	NoCompression      = zlib.NoCompression
	DefaultCompression = zlib.DefaultCompression
	BestSpeed          = zlib.BestSpeed
	BestCompression    = zlib.BestCompression
)

// Constants for NewBuffer
const (
	DefaultBuffer = 32 * 1024
)

// Errors specific to ZSeek
var (
	ErrEarlyWrite  = errors.New("zseek: cannot write before EOF")
	ErrInvalidSeek = errors.New("zseek: cannot seek outside of file")
)

type position struct {
	phys, virt int64
}

type ZSeek struct {
	f     io.ReadWriteSeeker
	read  bytes.Buffer
	write bytes.Buffer
	level int        // zlib compression level
	idx   []position // both values are monotonically increasing
	pos   position
	end   position // physical position is always set; virtual position is -1 until known
	buf   int      // max length of write before Flush is called automatically
	err   error
}

func New(f io.ReadWriteSeeker) (*ZSeek, error) {
	return NewLevel(f, DefaultCompression)
}

func NewLevel(f io.ReadWriteSeeker, level int) (*ZSeek, error) {
	return NewBuffer(f, level, DefaultBuffer)
}

func NewBuffer(f io.ReadWriteSeeker, level, buf int) (*ZSeek, error) {
	if buf <= 0 {
		buf = DefaultBuffer
	}

	end, err := f.Seek(0, SeekEnd)
	if err != nil {
		return nil, err
	}

	_, err = f.Seek(0, SeekStart)
	if err != nil {
		return nil, err
	}

	return &ZSeek{f: f, end: position{phys: end, virt: -1}, buf: buf, level: level}, nil
}

// Read implements io.Reader.
func (z *ZSeek) Read(p []byte) (n int, err error) {
	if z.err != nil {
		return 0, z.err
	}

	if z.read.Len() == 0 && len(p) != 0 {
		err = z.fill()
		if err != nil {
			return
		}
	}
	n, err = z.read.Read(p)
	z.pos.virt += int64(n)
	return
}

// Write implements io.Writer. Write can only be called when z is at the end of the file.
func (z *ZSeek) Write(p []byte) (n int, err error) {
	if z.err != nil {
		return 0, z.err
	}
	if z.read.Len() != 0 {
		return 0, ErrEarlyWrite
	}
	if z.write.Len() == 0 && z.pos.phys != z.end.phys {
		return 0, ErrEarlyWrite
	}

	z.write.Grow(len(p))
	for _, b := range p {
		err = z.write.WriteByte(b)
		if err != nil {
			return
		}
		z.pos.virt++
		n++

		if z.write.Len() >= z.buf {
			err = z.Flush()
			if err != nil {
				return
			}
		}
	}
	return
}

// Seek implements io.Seeker.
func (z *ZSeek) Seek(offset int64, whence int) (int64, error) {
	err := z.Flush()
	if err != nil {
		return 0, err
	}

	switch whence {
	case SeekEnd:
		if z.end.virt == -1 {
			err = z.seekEnd()
			if err != nil {
				return 0, err
			}
		}
		offset += z.end.virt

	case SeekCur:
		offset += z.pos.virt
	}

	if offset < 0 || (z.end.virt != -1 && z.end.virt < offset) {
		return 0, ErrInvalidSeek
	}

	// find the largest known position less than or equal to the offset we want
	i := sort.Search(len(z.idx), func(i int) bool {
		return z.idx[i].virt > offset
	}) - 1

	if i < 0 {
		z.pos = position{phys: 0, virt: 0}
	} else {
		z.pos = z.idx[i]
	}

	z.read.Reset()
	_, err = z.f.Seek(z.pos.phys, SeekStart)
	if err != nil {
		return 0, err
	}

	err = z.skip(offset - z.pos.virt)
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func (z *ZSeek) fill() error {
	var l int64
	err := binary.Read(z.f, binary.LittleEndian, &l)

	if err != nil {
		if err != io.EOF {
			z.err = err
		} else if z.pos.phys == z.end.phys {
			z.end.virt = z.pos.virt
		}
		return err
	}

	if len(z.idx) == 0 || z.idx[len(z.idx)-1].phys < z.pos.phys {
		z.idx = append(z.idx, z.pos)
	}

	z.pos.phys += l + (64 / 8)
	r, err := zlib.NewReader(io.LimitReader(z.f, l))
	if err != nil {
		z.err = err
		return err
	}
	_, err = io.Copy(&z.read, r)
	if err != nil {
		z.err = err
		return err
	}
	err = r.Close()
	if err != nil {
		z.err = err
		return err
	}
	return nil
}

func (z *ZSeek) seekEnd() error {
	for {
		z.pos.virt += int64(z.read.Len())
		z.read.Reset()
		err := z.fill()
		if err == io.EOF {
			if z.pos.phys == z.end.phys {
				return nil
			}
			z.err = io.ErrUnexpectedEOF
			return z.err
		} else if err != nil {
			return err
		}
	}
}

func (z *ZSeek) skip(n int64) error {
	_, err := io.CopyN(ioutil.Discard, z, n)
	if err != nil {
		z.err = err
		return err
	}
	return nil
}

// Flush writes any buffered data to the underlying io.ReadWriteSeeker. It is a no-op if
// there is no data to be written.
func (z *ZSeek) Flush() error {
	if z.err != nil {
		return z.err
	}
	toWrite := z.write.Len()
	if toWrite == 0 {
		return nil
	}
	var buf bytes.Buffer
	w, err := zlib.NewWriterLevel(&buf, z.level)
	if err != nil {
		z.err = err
		return err
	}
	vn, err := io.Copy(w, &z.write)
	if err != nil {
		z.err = err
		return err
	}
	if vn != int64(toWrite) {
		z.err = io.ErrShortWrite
		return z.err
	}
	err = w.Close()
	if err != nil {
		z.err = err
		return err
	}

	toWrite = buf.Len()

	const n int64 = 64 / 8
	err = binary.Write(z.f, binary.LittleEndian, int64(toWrite))
	if err != nil {
		z.err = err
		return err
	}

	pn, err := io.Copy(z.f, &buf)
	if err != nil {
		z.err = err
		return err
	}
	if pn != int64(toWrite) {
		z.err = io.ErrShortWrite
		return z.err
	}

	z.idx = append(z.idx, position{
		phys: z.pos.phys,
		virt: z.pos.virt - int64(vn),
	})
	z.pos.phys += n + pn
	z.end.phys += n + pn
	return nil
}

// Close implements io.Closer. Close does not close the underlying io.ReadWriteSeeker.
func (z *ZSeek) Close() error {
	err := z.Flush()
	z.err = io.ErrClosedPipe
	return err
}
