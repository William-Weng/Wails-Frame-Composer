package main

import (
	"embed"
	"frame-composer/backend"

	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
}

func main() {

	app := application.New(application.Options{
		Name:        "桌面圖片合成小工具",
		Description: "一個使用 Wails 3、Go、Svelte 與 Less 製作的桌面圖片合成工具",
		Services: []application.Service{
			application.NewService(&backend.FrameComposeService{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:          "桌面圖片合成小工具",
		Width:          500,
		Height:         600,
		EnableFileDrop: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
		},
		URL: "/",
	}).OnWindowEvent(events.Common.WindowFilesDropped, func(event *application.WindowEvent) {

		ctx := event.Context()
		files := ctx.DroppedFiles()
		target := ctx.DropTargetDetails()

		if len(files) > 0 {

			emitKey := "image-file-dropped"
			emitData := map[string]any{
				"slotId": target.ElementID,
				"path":   files[0],
			}

			application.Get().Event.Emit(emitKey, emitData)
		}
	})

	err := app.Run()

	if err != nil {
		log.Fatal(err)
	}
}
