package tools

// ImageType 表示目前要處理的圖片類型
type ImageType string

const (
	ImageTypeScreenshot ImageType = "screenshot" // ImageTypeScreenshot 表示內容截圖
	ImageTypeFrame      ImageType = "frame"      // ImageTypeFrame 表示外框圖片
)
