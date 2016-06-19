package img

type DummyProcessor struct{}

//Returns original data
func (p *DummyProcessor) Resize(data []byte, size string) ([]byte, error) {
	return data, nil
}

//Returns original data
func (p *DummyProcessor) FitToSize(data []byte, size string) ([]byte, error) {
	return data, nil
}
