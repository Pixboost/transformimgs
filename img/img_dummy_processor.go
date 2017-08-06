package img

type DummyProcessor struct{}

//Returns original data
func (p *DummyProcessor) Resize(data []byte, size string, imgId string) ([]byte, error) {
	return data, nil
}

//Returns original data
func (p *DummyProcessor) FitToSize(data []byte, size string, imgId string) ([]byte, error) {
	return data, nil
}

//Returns original data
func (p *DummyProcessor) Optimise(data []byte, imgId string) ([]byte, error) {
	return data, nil
}
