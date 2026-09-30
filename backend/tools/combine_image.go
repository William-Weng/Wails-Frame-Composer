package tools

import (
	"fmt"
	"image"
	"image/color"
	"os"
)

/* MARK: - 主程式 */
// 讀取並解碼指定路徑的圖片
//
// 參數：
//   - path：圖片檔案的路徑，例如 "frame.png"
//
// 回傳值：
//   - image.Image：解碼後的圖片；失敗時為 nil
//   - error：開啟或解碼失敗時的錯誤；成功時為 nil
func LoadImage(path string) (image.Image, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// 統計圖片中完全透明的像素數量
//
// 參數：
//   - img：要檢查的圖片
//
// 回傳值：
//   - transparent：alpha 為 0 的像素數量
//   - total：圖片的總像素數量
func InspectAlpha(img image.Image) (transparent, total int) {

	bounds := img.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			total++
			if alpha == 0 {
				transparent++
			}
		}
	}

	return
}

// 偵測手機外框中最大的封閉透明區域，並回傳該區域的矩形外接框
//
// 參數：
//   - frame：具有透明螢幕開孔的手機外框圖片
//
// 回傳值：
//   - image.Rectangle：螢幕透明區域的外接矩形；偵測失敗時為空矩形
//   - error：圖片尺寸無效或找不到封閉透明區域時的錯誤；成功時為 nil
func DetectScreenRegion(frame image.Image) (Region, error) {

	if err := checkSize(frame); err != nil {
		return Region{}, err
	}

	const alphaThreshold uint32 = 32 * 257
	transparent := transparentZone(frame, alphaThreshold)

	return findLargestEnclosedRegion(frame.Bounds(), transparent)
}

// 計算來源圖片的置中裁切範圍，讓裁切後的長寬比盡量符合目標區域
//
// 參數：
//   - src：來源圖片的座標範圍，例如 screenshot.Bounds()
//   - target：要填滿的目標區域，例如偵測出的螢幕矩形
//
// 回傳值：
//   - image.Rectangle：來源圖片中要保留的範圍。此函式只計算矩形；不會裁切圖片，也不會執行縮放
//
// 前提：src 和 target 的寬、高都必須大於 0
func CenterCrop(src, target image.Rectangle) image.Rectangle {

	srcWidth, srcHeight := src.Dx(), src.Dy()
	targetWidth, targetHeight := target.Dx(), target.Dy()

	// 用交叉相乘，避免浮點數運算
	if int64(srcWidth)*int64(targetHeight) > int64(srcHeight)*int64(targetWidth) {
		newWidth := int(int64(srcHeight) * int64(targetWidth) / int64(targetHeight))
		x := src.Min.X + (srcWidth-newWidth)/2
		return image.Rect(x, src.Min.Y, x+newWidth, src.Max.Y)
	}

	newHeight := int(int64(srcWidth) * int64(targetHeight) / int64(targetWidth))
	y := src.Min.Y + (srcHeight-newHeight)/2

	return image.Rect(src.Min.X, y, src.Max.X, y+newHeight)
}

// 將 BFS 找到的螢幕區域建立成 alpha mask
//
// 參數：
//   - frameBounds：外框圖片的座標範圍
//   - region：BFS 找到的螢幕透明區域
//   - width：外框圖片寬度，用來將一維索引轉換成 x、y 座標
//
// 回傳值：
//   - *image.Alpha：螢幕區域 mask。Alpha 為 255 的位置會顯示 screenshot；Alpha 為 0 的位置不會顯示 screenshot
func BuildScreenMask(frameBounds image.Rectangle, region Region, width int) *image.Alpha {

	mask := image.NewAlpha(frameBounds)

	for _, index := range region.pixels {
		x := frameBounds.Min.X + index%width
		y := frameBounds.Min.Y + index/width
		mask.SetAlpha(x, y, color.Alpha{A: 255})
	}

	return mask
}

/* MARK: - 小工具 */
// 檢查圖片的寬度與高度是否有效
//
// 參數：
//   - frame：要檢查尺寸的圖片
//
// 回傳值：
//   - error：寬度或高度小於等於 0 時回傳錯誤；有效時為 nil
func checkSize(frame image.Image) error {

	bounds := frame.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width <= 0 || height <= 0 {
		return fmt.Errorf("圖片尺寸無效")
	}

	return nil
}

// 依透明度門檻，標記圖片中的透明像素
//
// 參數：
//   - frame：要檢查的外框圖片。
//   - alphaThreshold：透明度門檻，使用 RGBA() 的 0～65535 尺度；alpha 小於或等於此值的像素會標記為 true
//
// 回傳值：
//   - []bool：長度為 width*height 的一維布林表。true 表示該像素符合透明條件；索引為 y*width+x，其中 x、y 是相對於 frame.Bounds().Min 的座標
func transparentZone(frame image.Image, alphaThreshold uint32) []bool {

	bounds := frame.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	transparent := make([]bool, width*height)

	for y := range height {
		for x := range width {
			_, _, _, alpha := frame.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			transparent[y*width+x] = alpha <= alphaThreshold
		}
	}

	return transparent
}

// 從透明像素表中，找出面積最大的封閉透明區域
//
// 參數：
//   - bounds：原始圖片的座標範圍，用來取得寬高，並將結果轉回原圖座標
//   - transparent：透明像素表，長度應為 bounds.Dx()*bounds.Dy()；索引 y*width+x 對應相對於 bounds.Min 的像素座標。true 表示該像素符合透明度門檻
//
// 回傳值：
//   - image.Rectangle：最大封閉透明區域的外接矩形；座標已轉換為原始圖片的座標系
//   - error：圖片尺寸、透明像素表長度不符，或找不到封閉區域時回傳錯誤
func findLargestEnclosedRegion(bounds image.Rectangle, transparent []bool) (Region, error) {

	width, height := bounds.Dx(), bounds.Dy()

	if width <= 0 || height <= 0 {
		return Region{}, fmt.Errorf("圖片尺寸無效")
	}

	if len(transparent) != width*height {
		return Region{}, fmt.Errorf("透明像素表長度錯誤")
	}

	visited := make([]bool, len(transparent))
	var best Region

	for start := range transparent {
		if !transparent[start] || visited[start] {
			continue
		}

		found := breadthFirstSearch(
			start, width, height, transparent, visited,
		)

		if found.touchesEdge || found.area <= best.area {
			continue
		}

		best = found
	}

	if best.area == 0 {
		return Region{}, fmt.Errorf("找不到封閉的透明螢幕區域")
	}

	// BFS 算出的 rect 是以 (0,0) 為原點的座標。
	// 轉成原圖 Bounds 的座標；pixels 仍保留一維相對索引。
	best.Rect = best.Rect.Add(bounds.Min)

	return best, nil
}

// 從指定像素開始，以廣度優先搜尋（BFS）找出與它上下左右相連的整塊透明區域
//
// 參數：
//   - start：起始像素在一維切片中的索引，計算方式為 y*width+x
//     呼叫前應確認 transparent[start] 為 true，且 visited[start] 為 false
//   - width：圖片寬度，用來在一維索引與 (x, y) 座標之間換算
//   - height：圖片高度，用來判斷像素是否位於圖片邊界
//   - transparent：透明像素地圖；true 表示該像素符合透明度門檻
//   - visited：走訪紀錄；函式會將找到的透明像素標記為 true，避免後續搜尋時重複處理同一塊區域
//
// 回傳值：
//   - Region：這塊相連透明區域的資訊，包含像素數量（area）、是否接觸圖片邊界（touchesEdge），以及外接矩形（rect）。rect 使用相對於圖片左上角 (0, 0) 的座標；若要轉成 frame.Bounds() 的座標，需再加上 bounds.Min
func breadthFirstSearch(start int, width, height int, transparent, visited []bool) Region {

	visited[start] = true
	queue := []int{start}

	minX, minY := width, height
	maxX, maxY := 0, 0

	result := Region{}

	for head := 0; head < len(queue); head++ {

		index := queue[head]
		x := index % width
		y := index / width

		result.area++
		result.pixels = append(result.pixels, index)

		if x == 0 || y == 0 || x == width-1 || y == height-1 {
			result.touchesEdge = true
		}

		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x+1 > maxX {
			maxX = x + 1
		}
		if y+1 > maxY {
			maxY = y + 1
		}

		neighbors := [][2]int{
			{x - 1, y},
			{x + 1, y},
			{x, y - 1},
			{x, y + 1},
		}

		for _, neighbor := range neighbors {
			nx, ny := neighbor[0], neighbor[1]

			if nx < 0 || nx >= width || ny < 0 || ny >= height {
				continue
			}

			next := ny*width + nx
			if transparent[next] && !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}

	result.Rect = image.Rect(minX, minY, maxX, maxY)

	return result
}
