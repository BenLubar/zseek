# zseek
--
    import "github.com/BenLubar/zseek"

Package zseek provides a seekable compressed file.

## Usage

```go
const (
	SeekStart = 0
	SeekCur   = 1
	SeekEnd   = 2
)
```
Constants for Seek

```go
const (
	NoCompression      = zlib.NoCompression
	DefaultCompression = zlib.DefaultCompression
	BestSpeed          = zlib.BestSpeed
	BestCompression    = zlib.BestCompression
)
```
Constants for NewLevel and NewBuffer

```go
const (
	DefaultBuffer = 32 * 1024
)
```
Constants for NewBuffer

```go
var (
	ErrEarlyWrite  = errors.New("zseek: cannot write before EOF")
	ErrInvalidSeek = errors.New("zseek: cannot seek outside of file")
)
```
Errors specific to ZSeek

#### type ZSeek

```go
type ZSeek struct {
}
```

ZSeek is a seekable compressed file. The file is written in chunks of
zlib-compressed data prefixed with a 64-bit little endian integer representing
the compressed size of the chunk. ZSeek can only be used as an io.Writer if it
is at the end of the file. Attempting to write before reaching the end of the
file will return ErrEarlyWrite.

#### func  New

```go
func New(f io.ReadWriteSeeker) (*ZSeek, error)
```
New is equivalent to calling NewBuffer(f, DefaultCompression, DefaultBuffer).

#### func  NewBuffer

```go
func NewBuffer(f io.ReadWriteSeeker, level, buf int) (*ZSeek, error)
```
NewBuffer creates a *ZSeek with a specified buffer size for writing. Whenever
there are at least buf bytes of unwritten data during a Write call, Flush will
automatically be called.

#### func  NewLevel

```go
func NewLevel(f io.ReadWriteSeeker, level int) (*ZSeek, error)
```
NewLevel is equivalent to calling NewBuffer(f, level, DefaultBuffer).

#### func (*ZSeek) Close

```go
func (z *ZSeek) Close() error
```
Close implements io.Closer. Close does not close the underlying
io.ReadWriteSeeker. After Close is called, any action on z will return
io.ErrClosedPipe.

#### func (*ZSeek) Flush

```go
func (z *ZSeek) Flush() error
```
Flush writes any buffered data to the underlying io.ReadWriteSeeker. Flush is a
no-op if there is no data to be written.

#### func (*ZSeek) Read

```go
func (z *ZSeek) Read(p []byte) (n int, err error)
```
Read implements io.Reader. If a read would cross a chunk boundary, a partial
read is done instead. Use io.ReadFull to guarantee a full read.

#### func (*ZSeek) Seek

```go
func (z *ZSeek) Seek(offset int64, whence int) (int64, error)
```
Seek implements io.Seeker. Flush will be called before an attempt is made to
seek.

#### func (*ZSeek) Write

```go
func (z *ZSeek) Write(p []byte) (n int, err error)
```
Write implements io.Writer. Write can only be called when z is at the end of the
file.
