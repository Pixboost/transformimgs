package img

type Image struct {
	Data     []byte
	MimeType string
}

// Info holds basic information about image
type Info struct {
	Format  string
	Quality int
	Opaque  bool
	Width   int
	Height  int
}
