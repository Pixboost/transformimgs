package img

type Image struct {
	Data     []byte
	MimeType string
}

// Info holds basic information about an image
type Info struct {
	Format  string
	Quality int
	Opaque  bool
	Width   int
	Height  int
	// Size is the size of the image in bytes
	Size int64
}
