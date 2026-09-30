package backend

type FrameComposeService struct{}

func (service *FrameComposeService) ImagePreview(path string) (string, error) {
	return imagePreview(path)
}

func (service *FrameComposeService) CombineImages(framePath string, screenshotPath string) (string, error) {
	return combineImages(framePath, screenshotPath)
}
