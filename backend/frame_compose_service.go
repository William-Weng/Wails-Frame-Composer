package backend

import (
	"frame-composer/backend/tools"
)

type FrameComposeService struct{}

// 將內容截圖裁切、縮放後放入外框偵測出的螢幕區域，再疊上外框，將結果存成 PNG
//
// 參數：
//   - framePath：外框圖片的完整檔案路徑
//   - screenshotPath：要放入外框的內容圖片完整檔案路徑
//
// 回傳值：
//   - 成功時回傳合成 PNG 的完整輸出路徑
//   - 讀取、偵測螢幕區域或寫入失敗時回傳 error
func (service *FrameComposeService) CombineImages(framePath string, screenshotPath string) (string, error) {
	return tools.CombineImages(framePath, screenshotPath)
}
