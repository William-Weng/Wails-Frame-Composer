package backend

import (
	"frame-composer/backend/tools"
	"io/fs"
	"net/http"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// 建立自訂的 HTTP Handler，負責處理媒體資源請求
//
// 參數:
//   - assets: Wails 的靜態資源 FS
//   - videoPreview: 影片預覽服務，提供可播放的影片路徑
//   - imagePreview: 圖片預覽服務，提供截圖與外框圖片路徑
//
// 回傳:
//   - http.Handler: 處理 /media/* 路徑的自訂 Handler，其餘路徑交由靜態資源處理
func NewAssetHandler(assets fs.FS, videoPreview *VideoPreviewService, imagePreview *ImagePreviewService) http.Handler {

	staticHandler := application.AssetFileServerFS(assets)

	videoInputPrefix := tools.MediaPrefix["video-input"] + "."
	screenshotPrefix := tools.MediaPrefix["image-screenshot"] + "."
	framePrefix := tools.MediaPrefix["image-frame"] + "."

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch {
		case strings.HasPrefix(request.URL.Path, videoInputPrefix):
			videoPreview.serveCurrentVideo(writer, request)
			return
		case strings.HasPrefix(request.URL.Path, screenshotPrefix):
			imagePreview.serveCurrentImage(writer, request, false)
			return
		case strings.HasPrefix(request.URL.Path, framePrefix):
			imagePreview.serveCurrentImage(writer, request, true)
			return
		default:
			staticHandler.ServeHTTP(writer, request)
		}
	})
}
