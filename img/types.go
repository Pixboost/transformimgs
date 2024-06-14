package img

type Image struct {
	// Id of the image mainly used for debugging purposes.
	// Could be a URL of the image or a filename.
	Id       string
	Data     []byte
	MimeType string
}

// Info holds basic information about an image.
type Info struct {
	Format  string
	Quality int
	Opaque  bool
	Width   int
	Height  int
	// Illustration is a flag set on PNG images.
	// If set to true then the image is an illustration, logo or
	// drawing and lossless compression would be preferable.
	// Otherwise, it's most likely a photo and lossy compression
	// could be used.
	Illustration bool
	// Size is the size of the image in bytes
	Size int64
}

// HttpError is user defined error that could be used for
// customising responses of the service
type HttpError struct {
	code int
	msg  string
}

func NewHttpError(code int, msg string) *HttpError {
	return &HttpError{code: code, msg: msg}
}

func (e *HttpError) Code() int {
	return e.code
}

func (e *HttpError) Error() string {
	return e.msg
}
