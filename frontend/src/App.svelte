<script lang="ts">
    import { onMount } from "svelte";
    import { Events } from "@wailsio/runtime";
    import { dialog } from "./utility/dialog";
    import { CombineImages } from "../bindings/frame-composer/backend/framecomposeservice"; // 換成實際產生的路徑
    import { SetFramePath, SetScreenshotPath } from "../bindings/frame-composer/backend/imagepreviewservice"; // 換成實際產生的路徑

    let leftImagePath = $state("");
    let rightImagePath = $state("");
    let leftImageSrc = $state("");
    let rightImageSrc = $state("");

    let outputPath = $state("");
    let isCombining = $state(false);

    // 每次拖入新圖或清除時遞增；用來忽略已過期的預覽結果
    let leftPreviewVersion = 0;
    let rightPreviewVersion = 0;

    onMount(() => {
        const unsubscribeDrop = Events.On("image-file-dropped", (event) => {
            void imageFileDroppedAction(event.data as ImageDroppedData);
        });

        return () => unsubscribeDrop();
    });

    /**
     * 處理從檔案系統拖入圖片的事件
     *
     * @param data - Wails 拖放事件資料，包含目標圖框 ID 與圖片完整路徑
     * @returns 圖片預覽處理完成後結束；不回傳資料
     */
    async function imageFileDroppedAction(
        data: ImageDroppedData,
    ): Promise<void> {
        if (data.slotId !== "leftSlot" && data.slotId !== "rightSlot") {
            return;
        }

        const isLeft = data.slotId === "leftSlot";
        const version = isLeft
            ? ++leftPreviewVersion
            : ++rightPreviewVersion;

        try {
            // 左側是外框圖，右側是內容圖。
            // Go service 會驗證路徑、記住目前檔案，
            // 並回傳可供 WebView 使用的虛擬 URL。
            const src = isLeft
                ? await SetFramePath(data.path)
                : await SetScreenshotPath(data.path);

            // 如果等待後使用者已拖入另一張圖，忽略舊請求結果。
            if (isLeft && version !== leftPreviewVersion) {
                return;
            }

            if (!isLeft && version !== rightPreviewVersion) {
                return;
            }

            if (isLeft) {
                leftImagePath = data.path;
                leftImageSrc = src;
            } else {
                rightImagePath = data.path;
                rightImageSrc = src;
            }

            outputPath = "";
        } catch (err) {
            // 已經有更新的拖放請求時，不顯示舊請求的錯誤。
            if (isLeft && version !== leftPreviewVersion) {
                return;
            }

            if (!isLeft && version !== rightPreviewVersion) {
                return;
            }

            const error = err instanceof Error
                ? err.message
                : String(err)

            await dialog("warning", "圖片預覽失敗", error);
        }
    }

    /**
     * 使用目前選取的外框圖與內容圖進行合成
     *
     * @returns 合成與結果提示結束後完成；不回傳資料
     */
    async function combineImage(): Promise<void> {

        if (!leftImagePath || !rightImagePath || isCombining) { return; }

        isCombining = true;

        try {
            outputPath = await CombineImages(leftImagePath, rightImagePath);
            isCombining = false;
            await dialog("info", "合成完成", outputPath);
        } catch (err) {
          const error = err instanceof Error ? err.message : String(err);
          isCombining = false;
          await dialog("warning", "圖片合成失敗", error);
        }
    }

    /**
     * 清除兩個圖框的圖片、檔案路徑與合成結果
     */
    function resetInputPath(): void {
        leftPreviewVersion++;
        rightPreviewVersion++;
        leftImagePath = "";
        rightImagePath = "";
        leftImageSrc = "";
        rightImageSrc = "";
        outputPath = "";
    }
</script>

<main>
    <section id="leftSlot" class="image-slot" data-file-drop-target>
        <div class="image-slot-frame">
            {#if leftImageSrc}
                <img class="image-slot-frame-img" src={leftImageSrc} alt="外框圖預覽">
            {:else}
                <span class="placeholder">外框圖預覽</span>
            {/if}
        </div>
        <div id="rightSlot" class="image-slot-frame" data-file-drop-target>
            {#if rightImageSrc}
                <img class="image-slot-frame-img" src={rightImageSrc} alt="內容圖預覽">
            {:else}
                <span class="placeholder">內容圖預覽</span>
            {/if}
        </div>
    </section>
    <section class="button-area">
        <button id="composeButton" class="button-area-frame" disabled={!leftImagePath || !rightImagePath || isCombining} onclick={combineImage}>
            {isCombining ? "合成中…" : "合成"}
        </button>
        <button id="clearButton" class="button-area-frame" disabled={!leftImagePath && !rightImagePath || isCombining} onclick={resetInputPath}>
            清除
        </button>
    </section>
</main>
