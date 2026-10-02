package backend

import (
	"frame-composer/backend/tools"
	"net/http"
	"path/filepath"
	"strings"

	"sync"
)

type ImagePreviewService struct {
	mutex          sync.RWMutex
	screenshotPath string
	framePath      string
}

// 設定目前要在前端預覽的外框圖片，回傳的 URL 可直接指定給 <img src>
//
// 參數:
//   - path: 圖片檔案路徑（相對或絕對路徑）
//
// 回傳:
//   - url: 前端可用的虛擬 URL（例如 "/media/image/frame.png?v=123"）
//   - error: 若路徑無效或檔案不存在則回傳錯誤
func (service *ImagePreviewService) SetFramePath(path string) (string, error) {
	prefix := tools.MediaPrefix["image-frame"]
	return service.setImagePath(path, prefix, true)
}

// 設定目前要在前端預覽的截圖圖片，回傳的 URL 可直接指定給 <img src>
//
// 參數:
//   - path: 圖片檔案路徑（相對或絕對路徑）
//
// 回傳:
//   - url: 前端可用的虛擬 URL（例如 "/media/image/screenshot.png?v=123"）
//   - error: 若路徑無效或檔案不存在則回傳錯誤
func (service *ImagePreviewService) SetScreenshotPath(path string) (string, error) {
	prefix := tools.MediaPrefix["image-screenshot"]
	return service.setImagePath(path, prefix, false)
}

// 內部共用函式，負責驗證並設定圖片路徑
//
// 參數:
//   - path: 圖片檔案路徑
//   - prefix: 對應的 URL 前綴（來自 mediaPrefix）
//
// 回傳:
//   - url: 組合後的虛擬 URL
//   - error: 驗證失敗時的錯誤
func (service *ImagePreviewService) setImagePath(path string, prefix string, isFrame bool) (string, error) {

	absolutePath, info, err := tools.ValidatePath(path, tools.ImageMimeMap)
	if err != nil {
		return "", err
	}

	service.mutex.Lock()
	if isFrame {
		service.framePath = absolutePath
	} else {
		service.screenshotPath = absolutePath
	}
	service.mutex.Unlock()

	return tools.MediaURL(prefix, absolutePath, info), nil
}

// serveCurrentImage 處理圖片預覽請求，根據 isFrame 決定回傳截圖或外框。
//
// 參數:
//   - writer: HTTP 回應寫入器
//   - request: HTTP 請求物件
//   - service: 圖片預覽服務
//   - isFrame: true 回傳外框圖片，false 回傳截圖
func (service *ImagePreviewService) serveCurrentImage(writer http.ResponseWriter, request *http.Request, isFrame bool) {

	service.mutex.RLock()
	path := service.screenshotPath
	if isFrame {
		path = service.framePath
	}
	service.mutex.RUnlock()

	file, err := tools.OpenFile(path)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer file.Close()

	contentType, ok := tools.ImageContentType(strings.ToLower(filepath.Ext(path)))
	if !ok {
		http.Error(writer, "不支援的圖片格式", http.StatusUnsupportedMediaType)
		return
	}

	tools.ServeFile(writer, request, file, contentType)
}
