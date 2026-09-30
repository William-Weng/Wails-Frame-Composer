package backend

import (
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/http"
	"os"
	"strings"

	"frame-composer/backend/tools"

	"golang.org/x/image/draw"
)

// 讀取本機圖片，並轉成可供 HTML <img src> 使用的 Data URL
//
// 參數：
//   - path：圖片檔案的完整路徑，例如 /Users/ios/Desktop/古本屋.jpg
//
// 回傳值：
//   - 成功時回傳 data:image/...;base64,... 格式的字串
//   - 檔案讀取失敗或檔案內容不是圖片時，回傳 error
func imagePreview(path string) (string, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("不是支援的圖片檔案: %s", path)
	}

	return "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

// 將內容截圖裁切、縮放後放入外框偵測出的螢幕區域，再疊上外框，將結果存成 PNG
//
// 參數：
//   - framePath：外框圖片的完整檔案路徑
//   - screenshotPath：要放入外框的內容圖片完整檔案路徑
//
// 回傳值：
//   - 成功時回傳合成 PNG 的完整輸出路徑
//   - 讀取、偵測螢幕區域或寫入失敗時回傳 error
func combineImages(framePath string, screenshotPath string) (string, error) {

	frame, err := tools.LoadImage(framePath)
	if err != nil {
		return "", err
	}

	screenshot, err := tools.LoadImage(screenshotPath)
	if err != nil {
		return "", err
	}

	// transparent, total := tools.InspectAlpha(frame)
	// fmt.Printf("透明像素：%d / %d，比例：%.2f%%\n", transparent, total, float64(transparent)/float64(total)*100)

	region, err := tools.DetectScreenRegion(frame)

	if err != nil {
		return "", err
	}

	screenRect := region.Rect

	// fmt.Printf("偵測到螢幕區域：%v\n", screenRect)
	// fmt.Printf("外框尺寸：%v\n", frame.Bounds())
	frameBounds := frame.Bounds()

	// 輸出畫布，尺寸完全跟 frame.png 一樣
	result := image.NewRGBA(frameBounds)
	sourceRect := tools.CenterCrop(screenshot.Bounds(), screenRect)

	screenWidth := screenRect.Dx()
	screenHeight := screenRect.Dy()

	// 先把 screenshot 等比例縮放到螢幕矩形
	scaled := image.NewRGBA(image.Rect(0, 0, screenWidth, screenHeight))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), screenshot, sourceRect, draw.Src, nil)

	// 螢幕圓角半徑 (BFS)
	mask := tools.BuildScreenMask(frame.Bounds(), region, frame.Bounds().Dx())

	// 截圖從 scaled 的 (0,0) 讀取；mask 則從 screenRect.Min 讀取；兩者都對齊 result 的 screenRect.Min
	draw.DrawMask(result, screenRect, scaled, scaled.Bounds().Min, mask, screenRect.Min, draw.Src)

	// 外框最後疊上，保留邊框與 Dynamic Island
	draw.Draw(result, frameBounds, frame, frameBounds.Min, draw.Over)

	outputPath := tools.OutputPath(screenshotPath)
	output, err := os.Create(outputPath)

	if err != nil {
		return "", err
	}

	defer output.Close()

	if err := png.Encode(output, result); err != nil {
		return "", err
	}

	return outputPath, nil
}
