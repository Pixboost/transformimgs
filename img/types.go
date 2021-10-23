package img

type Image struct {
	// Id of the image mainly used for debugging purposes
	Id       string
	Data     []byte
	MimeType string
}

// Info holds basic information about an image
type Info struct {
	Format       string
	Quality      int
	Opaque       bool
	Width        int
	Height       int
	Illustration bool
	// Size is the size of the image in bytes
	Size int64
}
