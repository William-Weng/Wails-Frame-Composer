# [手機邊框組合小工具](https://github.com/William-Weng?tab=repositories&q=wails)

一個使用 Wails 3、Go、Svelte 與 Less 製作的桌面圖片合成工具。將手機外框圖與內容截圖拖進左右圖框，程式會偵測外框中的螢幕區域，把截圖放入後輸出 PNG。

https://github.com/user-attachments/assets/73a9a678-4c18-4665-a0f8-be13986846d6

## [功能](https://peterpanswift.github.io/iphone-bezels/)

- 從系統檔案管理器將圖片拖到「外框」或「內容」圖框。
- 在圖框內置中預覽圖片，保持原始比例，不拉伸。
- 偵測外框圖片中的螢幕區域，將內容截圖置中裁切、縮放並套用螢幕遮罩。
- 最後疊上外框，保留邊框等前景細節，輸出與外框相同尺寸的 PNG。
- 合成完成或失敗時顯示系統對話框；「清除」可重設目前輸入。

> 外框圖片需要有可供程式辨識的透明螢幕區域。若偵測不到螢幕區域，合成會失敗。此工具目前的預覽由 Go 讀取圖片並傳回 Base64 Data URL；輸出圖片則直接寫入檔案。

## [使用方式](https://www.flaticon.com/free-icon/responsive-design_8488732)

1. 將手機外框圖片拖入左側「外框」圖框。
2. 將要放進螢幕的截圖拖入右側「內容」圖框。
3. 確認兩張圖片的預覽，按下「合成」。
4. 在完成對話框中查看輸出檔案路徑；輸出檔案位於內容截圖所在的目錄，檔名帶有時間戳並以 `.png` 結尾。
5. 按「清除」可移除目前的兩張預覽與路徑。

預覽圖使用完整顯示模式：當圖片與圖框的長寬比不同時，周圍留空是正常現象；這不會將圖片拉伸變形。

## 開發

專案使用 Go 作為圖片處理與 Wails 後端，Svelte／TypeScript 管理拖放事件與介面狀態，Less 管理深色介面樣式。安裝專案所需的 Go、Node.js 與 Wails 3 工具後，在專案根目錄執行：

```sh
cd frontend
npm install
cd ..
wails3 dev
```

建置指令與產物位置請以專案的 `Taskfile.yml` 及 Wails 設定為準；不同 Wails 3 專案模板可能有不同的建置 task。

## 工作流程

```text
拖入外框圖 / 內容圖
    ↓
Wails 取得檔案路徑及圖框 ID（leftSlot / rightSlot）
    ↓
Svelte 接收 image-file-dropped 事件並呼叫 ImagePreview
    ↓
Go 讀取圖片，回傳預覽用 Data URL
    ↓
按下「合成」→ 呼叫 CombineImages(外框路徑, 內容路徑)
    ↓
偵測螢幕區域 → 裁切／縮放內容 → 套用遮罩 → 疊上外框
    ↓
寫入 PNG，回傳輸出路徑
```

## 技術與注意事項

- **Wails 3**：橋接 Go 與前端，並接收系統檔案拖放事件。
- **Go**：讀取圖片、偵測透明區域與螢幕範圍、合成並寫入 PNG。
- **Svelte／TypeScript**：顯示預覽、管理左右圖框狀態與合成操作。
- **Less**：管理深色配色、圖片區域與固定高度按鈕。
- 預覽使用 Data URL，較大的圖片會增加前後端傳輸與記憶體使用量；若需要處理大量或高解析度圖片，可改用 Wails 資源處理器提供預覽 URL。
- 合成結果與畫面上的預覽是兩件事：目前成功對話框顯示輸出檔案路徑，並未在介面中另設結果預覽框。

## [建置指令整理](https://v3.wails.io/zh-tw/guides/build/building/)

| 目的 | 指令 |
| --- | --- |
| 建立新專案 | `wails3 init -n <專案名稱> -t <前端框架>` |
| 產生 bindings | `wails3 generate bindings` |
| 更新 Windows 與 macOS 專用圖示檔案 | `wails3 generate icons -input build/appicon.png -windowsfilename build/windows/icon.ico -macfilename build/darwin/icons.icns` |
| 更新建置資源 | `wails3 update build-assets -config build/config.yml -dir build` |
| 拉取（下載）用於跨平台交叉編譯的 Docker 映像檔 | `wails3 task setup:docker` |
| 建置 Windows x64 | `wails3 build GOOS=windows GOARCH=amd64` |
| 打包 macOS arm64 | `wails3 package GOOS=darwin GOARCH=arm64` |
| 打包 Linux arm64 | `wails3 build GOOS=linux GOARCH=arm64` |

## 專案檔案

```text
.
├── main.go                 # Wails 應用入口與拖放事件設定
├── frontend/
│   └── src/
│       ├── App.svelte      # 圖框、預覽、合成與清除操作
│       └── main.less       # 深色介面樣式
├── go.mod                  # Go 模組設定
└── Taskfile.yml            # 專案建置 task（依實際專案配置）
```
