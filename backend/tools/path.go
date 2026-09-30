package tools

import (
	"path/filepath"
	"strings"
	"time"
)

// 根據輸入圖片路徑，產生同目錄下的 PNG 輸出路徑
//
// 參數：
//   - inputPath：輸入圖片的完整路徑，例如 /Users/ios/Desktop/古本屋.jpg
//
// 回傳值：
//   - 輸出圖片的完整路徑。檔名會移除原副檔名，加入目前時間，並以 .png 結尾，例如 /Users/ios/Desktop/古本屋-20260930-134605.png
//
// 此方法只產生路徑，不會建立或寫入圖片檔案
func OutputPath(inputPath string) string {

	ext := filepath.Ext(inputPath)
	basePath := strings.TrimSuffix(inputPath, ext)
	timestamp := time.Now().Format("20060102-150405")

	return basePath + "-" + timestamp + ".png"
}
