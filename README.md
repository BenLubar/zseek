# zseek
--
    import "github.com/BenLubar/zseek"


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


#### func  New

```go
func New(f io.ReadWriteSeeker) (*ZSeek, error)
```

#### func  NewBuffer

```go
func NewBuffer(f io.ReadWriteSeeker, level, buf int) (*ZSeek, error)
```

#### func  NewLevel

```go
func NewLevel(f io.ReadWriteSeeker, level int) (*ZSeek, error)
```

#### func (*ZSeek) Close

```go
func (z *ZSeek) Close() error
```
Close implements io.Closer. Close does not close the underlying
io.ReadWriteSeeker.

#### func (*ZSeek) Flush

```go
func (z *ZSeek) Flush() error
```
Flush writes any buffered data to the underlying io.ReadWriteSeeker. It is a
no-op if there is no data to be written.

#### func (*ZSeek) Read

```go
func (z *ZSeek) Read(p []byte) (n int, err error)
```
Read implements io.Reader.

#### func (*ZSeek) Seek

```go
func (z *ZSeek) Seek(offset int64, whence int) (int64, error)
```
Seek implements io.Seeker.

#### func (*ZSeek) Write

```go
func (z *ZSeek) Write(p []byte) (n int, err error)
```
Write implements io.Writer. Write can only be called when z is at the end of the
file.
