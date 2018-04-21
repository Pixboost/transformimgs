package processors

type Dummy struct{}

//Returns original data
func (p *Dummy) Resize(data []byte, size string, imgId string) ([]byte, error) {
	return data, nil
}

//Returns original data
func (p *Dummy) FitToSize(data []byte, size string, imgId string) ([]byte, error) {
	return data, nil
}

//Returns original data
func (p *Dummy) Optimise(data []byte, imgId string, supportedFormats []string) ([]byte, error) {
	return data, nil
}
