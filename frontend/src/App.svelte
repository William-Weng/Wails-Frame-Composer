<script lang="ts">
    import { onMount } from "svelte";
    import { Events } from "@wailsio/runtime";
    import { dialog } from "./utility/dialog";
    import { ImagePreview, CombineImages } from "../bindings/frame-composer/backend/framecomposeservice"; // 換成實際產生的路徑

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
    async function imageFileDroppedAction(data: ImageDroppedData): Promise<void> {

        if (data.slotId !== "leftSlot" && data.slotId !== "rightSlot") { return; }

        const isLeft = data.slotId === "leftSlot";
        const version = isLeft ? ++leftPreviewVersion : ++rightPreviewVersion;

        try {
            const src = await ImagePreview(data.path);

            // 期間如果又拖入另一張圖或按了清除，就忽略舊結果
            if (isLeft && version !== leftPreviewVersion) return;
            if (!isLeft && version !== rightPreviewVersion) return;

            if (isLeft) {
                leftImagePath = data.path;
                leftImageSrc = src;
            } else {
                rightImagePath = data.path;
                rightImageSrc = src;
            }

            outputPath = "";
        } catch (err) {
            // 失敗了，但如果已是過期請求，就不要顯示錯誤對話框
            if (isLeft && version !== leftPreviewVersion) { return; }
            if (!isLeft && version !== rightPreviewVersion) { return; }
            const error = err instanceof Error ? err.message : String(err);
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
            await dialog("info", "合成完成", outputPath);
        } catch (err) {
          const error = err instanceof Error ? err.message : String(err);
          await dialog("warning", "圖片合成失敗", error);
        } finally {
            isCombining = false;
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
    <div class="image-area">
        <section id="leftSlot" class="image-slot" class:has-image={!!leftImageSrc} data-file-drop-target>
            {#if leftImageSrc}
                <img src={leftImageSrc} alt="外框預覽" />
            {:else}
                <span class="placeholder">外框</span>
            {/if}
        </section>
        <section id="rightSlot" class="image-slot" class:has-image={!!rightImageSrc} data-file-drop-target>
            {#if rightImageSrc}
                <img src={rightImageSrc} alt="內容圖預覽" />
            {:else}
                <span class="placeholder">內容</span>
            {/if}
        </section>
    </div>

    <div class="actions">
        <button id="composeButton" type="button" onclick={combineImage} disabled={!leftImagePath || !rightImagePath || isCombining}>
            {isCombining ? "合成中…" : "合成"}
        </button>
        <button id="clearButton" type="button" onclick={resetInputPath}>
            清除
        </button>
    </div>
</main>
