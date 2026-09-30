package tools

import "image"

// Region 表示一塊相連透明區域的搜尋結果。Rect 的座標相對於圖片左上角；Max 不包含在矩形內
type Region struct {
	Rect        image.Rectangle
	area        int
	touchesEdge bool
	pixels      []int
}
