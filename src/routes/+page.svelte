<script lang="ts">
  import { onMount, onDestroy } from "svelte";

  type ImageItem = {
    name: string;
    url: string; // object URL
    size: number;
    lastModified: number;
    zipPath?: string;
    zipEntry?: string;
  };

  const IMAGE_EXTS = [
    ".jpg",
    ".jpeg",
    ".png",
    ".gif",
    ".webp",
    ".avif",
    ".bmp"
  ];

  let filesInput: HTMLInputElement | null = null;
  let images: ImageItem[] = $state([]);
  let current = $state(0);
  let fitWidth = $state(true);
  let zoom = $state(1);
  let twoPage = $state(false);
  let rtl = $state(false); // 右開き(右→左)
  let isZipMode = $state(false);

  let stageEl: HTMLElement | null = null;
  let wheelLock = false;

  async function openFolderPicker() {
    // Prefer native dialog in Tauri; fallback to browser folder input
    const isTauri = typeof window !== "undefined" && "__TAURI__" in window;
    if (isTauri) {
      try {
        const { invoke, convertFileSrc } = await import("@tauri-apps/api/core");
        const dir = (await invoke("plugin:dialog|open", { directory: true, multiple: false })) as string | string[] | null;
        if (!dir || Array.isArray(dir)) return;

        const paths: string[] = await invoke("list_images_in_dir", { dir });

        // Revoke old blob URLs only
        for (const it of images) {
          if (it.url.startsWith("blob:")) URL.revokeObjectURL(it.url);
        }

        const items: ImageItem[] = paths.map((p) => {
          const name = p.split(/[/\\]/).pop() || p;
          return {
            name,
            url: convertFileSrc(p),
            size: 0,
            lastModified: 0
          };
        });
        // Client-side natural sort by name
        items.sort((a, b) => naturalCompare(a.name, b.name));
        images = items;
        current = 0;
        return;
      } catch (err) {
        console.warn("Tauri dialog failed; fallback to browser picker", err);
      }
    }
    // Fallback: browser folder input
    filesInput?.click();
  }

  async function openZipPicker() {
    const isTauri = typeof window !== "undefined" && "__TAURI__" in window;
    if (isTauri) {
      try {
        const { invoke } = await import("@tauri-apps/api/core");
        const file = (await invoke("plugin:dialog|open", {
          directory: false,
          multiple: false,
          filters: [{ name: "ZIP/CBZ", extensions: ["zip", "cbz"] }]
        })) as string | string[] | null;
        if (!file || Array.isArray(file)) return;

        const entries: string[] = await invoke("list_images_in_zip", { path: file });
        entries.sort((a, b) => naturalCompare(a, b));

        // revoke only blob URLs
        for (const it of images) {
          if (it.url.startsWith("blob:")) URL.revokeObjectURL(it.url);
        }

        images = entries.map((entry) => ({
          name: entry,
          url: "",
          size: 0,
          lastModified: 0,
          zipPath: file,
          zipEntry: entry
        }));
        isZipMode = true;
        current = 0;
        await ensureImageLoaded(current);
        if (twoPage) {
          const n2 = rtl ? current - 1 : current + 1;
          await ensureImageLoaded(n2);
        }
        return;
      } catch (err) {
        console.error("Tauri zip open failed", err);
      }
    }
    alert("ブラウザ開発モードではZIP読み込みは未対応です。Tauriでお試しください。");
  }

  async function ensureImageLoaded(index: number) {
    if (index < 0 || index >= images.length) return;
    const it = images[index];
    if (!it) return;
    if (it.url && it.url.length > 0) return;
    if (!it.zipPath || !it.zipEntry) return;
    try {
      const { invoke } = await import("@tauri-apps/api/core");
      const [mime, b64]: [string, string] = await invoke("read_zip_image", {
        path: it.zipPath,
        entry: it.zipEntry
      });
      it.url = `data:${mime};base64,${b64}`;
      images = images.slice();
    } catch (e) {
      console.error("read_zip_image failed", e);
    }
  }

  $effect(() => {
    if (!isZipMode || images.length === 0) return;
    ensureImageLoaded(current);
    if (twoPage) {
      const n2 = rtl ? current - 1 : current + 1;
      ensureImageLoaded(n2);
    }
  });

  function naturalCompare(a: string, b: string) {
    // human-friendly: 1 < 2 < 10
    return a.localeCompare(b, undefined, { numeric: true, sensitivity: "base" });
  }

  function onPickFiles(e: Event) {
    const input = e.target as HTMLInputElement;
    const list = input.files;
    if (!list || list.length === 0) return;

    // Revoke old URLs
    for (const it of images) URL.revokeObjectURL(it.url);

    const items: ImageItem[] = [];
    for (const file of Array.from(list)) {
      const lower = file.name.toLowerCase();
      if (!IMAGE_EXTS.some((ext) => lower.endsWith(ext))) continue;
      items.push({
        name: file.webkitRelativePath || file.name,
        url: URL.createObjectURL(file),
        size: file.size,
        lastModified: file.lastModified
      });
    }

    items.sort((a, b) => naturalCompare(a.name, b.name));
    images = items;
    current = 0;
  }

  function next() {
    if (images.length === 0) return;
    current = Math.min(current + 1, images.length - 1);
  }
  function prev() {
    if (images.length === 0) return;
    current = Math.max(current - 1, 0);
  }
  function clampIndex(i: number) {
    return Math.max(0, Math.min(i, images.length - 1));
  }
  function jumpBy(delta: number) {
    if (images.length === 0) return;
    current = clampIndex(current + delta);
  }
  function forward() {
    // 読み方向に進む
    const step = twoPage ? 2 : 1;
    jumpBy(rtl ? -step : step);
  }
  function backward() {
    // 読み方向に戻る
    const step = twoPage ? 2 : 1;
    jumpBy(rtl ? step : -step);
  }
  function first() {
    if (images.length === 0) return;
    current = 0;
  }
  function last() {
    if (images.length === 0) return;
    current = images.length - 1;
  }

  function onKey(e: KeyboardEvent) {
    switch (e.key) {
      case "ArrowRight":
      case "PageDown":
      case "]":
      case " ":
        e.preventDefault();
        forward();
        break;
      case "ArrowLeft":
      case "PageUp":
      case "[":
        e.preventDefault();
        backward();
        break;
      case "Home":
        e.preventDefault();
        first();
        break;
      case "End":
        e.preventDefault();
        last();
        break;
      case "+":
        e.preventDefault();
        zoom = Math.min(zoom + 0.1, 4);
        break;
      case "-":
        e.preventDefault();
        zoom = Math.max(zoom - 0.1, 0.2);
        break;
      case "0":
        e.preventDefault();
        zoom = 1;
        break;
      case "f":
      case "F":
        // 幅合わせトグル
        e.preventDefault();
        fitWidth = !fitWidth;
        break;
      case "d":
      case "D":
        // 2ページ表示トグル
        e.preventDefault();
        twoPage = !twoPage;
        break;
      case "r":
      case "R":
        // 右開きトグル
        e.preventDefault();
        rtl = !rtl;
        break;
    }
  }

  function onWheel(e: WheelEvent) {
    // Ctrl+ホイールはブラウザ拡大縮小と競合するので無視
    if (e.ctrlKey) return;
    if (wheelLock) return;
    wheelLock = true;
    const dir = e.deltaY > 0 ? 1 : -1;
    if (dir > 0) forward(); else backward();
    // 連続スクロールでページ送りしすぎないよう少し間を空ける
    setTimeout(() => (wheelLock = false), 120);
  }

  // ドラッグでスクロール（パン）
  let dragging = false;
  let lastX = 0, lastY = 0;
  let startScrollLeft = 0, startScrollTop = 0;
  function onPointerDown(e: PointerEvent) {
    if (!stageEl) return;
    dragging = true;
    stageEl.setPointerCapture(e.pointerId);
    lastX = e.clientX;
    lastY = e.clientY;
    startScrollLeft = stageEl.scrollLeft;
    startScrollTop = stageEl.scrollTop;
  }
  function onPointerMove(e: PointerEvent) {
    if (!dragging || !stageEl) return;
    const dx = e.clientX - lastX;
    const dy = e.clientY - lastY;
    stageEl.scrollLeft = startScrollLeft - dx;
    stageEl.scrollTop = startScrollTop - dy;
  }
  function onPointerUp(e: PointerEvent) {
    if (!stageEl) return;
    dragging = false;
    try { stageEl.releasePointerCapture(e.pointerId); } catch {}
  }

  function onStageClick(e: MouseEvent) {
    // 左右クリックでページ送り/戻り
    if (!stageEl) return;
    const rect = stageEl.getBoundingClientRect();
    const leftHalf = e.clientX - rect.left < rect.width / 2;
    if (leftHalf) backward(); else forward();
  }

  onMount(() => {
    window.addEventListener("keydown", onKey, { passive: false });
  });
  onDestroy(() => {
    window.removeEventListener("keydown", onKey);
    for (const it of images) URL.revokeObjectURL(it.url);
  });
</script>

<main class="viewer-root">
  <header class="toolbar">
    <button onclick={openFolderPicker}>フォルダを開く</button>
    <button onclick={openZipPicker}>ZIP/CBZを開く</button>
    <input
      bind:this={filesInput}
      type="file"
      webkitdirectory
      multiple
      onchange={onPickFiles}
      class="hidden-input"
    />

    <div class="spacer"></div>

    <button onclick={first} disabled={images.length === 0 || current === 0}>
      «
    </button>
    <button onclick={backward} disabled={images.length === 0 || current === 0}>
      ‹
    </button>
    <span class="page-info"
      >{images.length === 0 ? "0 / 0" : `${current + 1} / ${images.length}`}</span
    >
    <button
      onclick={forward}
      disabled={images.length === 0 || current === images.length - 1}
    >
      ›
    </button>
    <button
      onclick={last}
      disabled={images.length === 0 || current === images.length - 1}
    >
      »
    </button>

    <div class="divider"></div>
    <label class="toggle">
      <input type="checkbox" bind:checked={fitWidth} /> 幅に合わせる
    </label>
    <label class="toggle">
      <input type="checkbox" bind:checked={twoPage} /> 2ページ表示
    </label>
    <label class="toggle">
      <input type="checkbox" bind:checked={rtl} /> 右開き
    </label>
    <div class="zoom">
      <input
        type="range"
        min="0.2"
        max="4"
        step="0.1"
        bind:value={zoom}
      />
      <span>{Math.round(zoom * 100)}%</span>
    </div>
  </header>

  {#if images.length === 0}
    <section class="empty">
      <img src="/tauri.svg" alt="logo" />
      <h2>フォルダを選んで読み込み</h2>
      <p>画像ファイル（jpg, png, webp など）を含むフォルダを選択してください。</p>
      <div style="display:flex; gap:8px;">
        <button onclick={openFolderPicker}>フォルダを開く</button>
        <button onclick={openZipPicker}>ZIP/CBZを開く</button>
      </div>
    </section>
  {:else}
    <section
      class="stage {fitWidth ? 'fit-width' : ''} {twoPage ? 'two' : ''}"
      bind:this={stageEl}
      onwheel={onWheel}
      onpointerdown={onPointerDown}
      onpointermove={onPointerMove}
      onpointerup={onPointerUp}
      onclick={onStageClick}
      ondblclick={() => (fitWidth = !fitWidth)}
      role="group"
      tabindex="0"
      onkeydown={onKey}
    >
      {#if twoPage}
        {#if rtl}
          {#if current - 1 >= 0}
            <img alt={images[current].name} src={images[current].url} style={`transform: scale(${zoom});`} />
            <img alt={images[current - 1].name} src={images[current - 1].url} style={`transform: scale(${zoom});`} />
          {:else}
            <img alt={images[current].name} src={images[current].url} style={`transform: scale(${zoom});`} />
          {/if}
        {:else}
          {#if current + 1 < images.length}
            <img alt={images[current].name} src={images[current].url} style={`transform: scale(${zoom});`} />
            <img alt={images[current + 1].name} src={images[current + 1].url} style={`transform: scale(${zoom});`} />
          {:else}
            <img alt={images[current].name} src={images[current].url} style={`transform: scale(${zoom});`} />
          {/if}
        {/if}
      {:else}
        <img alt={images[current].name} src={images[current].url} style={`transform: scale(${zoom});`} />
      {/if}
      <div class="filename">{images[current].name}</div>
    </section>
  {/if}
</main>

<style>
  :root {
    font-family: Inter, Avenir, Helvetica, Arial, sans-serif;
    font-size: 16px;
    color: var(--fg, #0f0f0f);
    background-color: var(--bg, #f6f6f6);
  }

  .viewer-root {
    display: grid;
    grid-template-rows: auto 1fr;
    height: 100svh;
  }

  .toolbar {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid #ddd;
    background: #fff8;
    backdrop-filter: blur(6px);
  }
  .toolbar button {
    border-radius: 6px;
    border: 1px solid #c9c9c9;
    background: #fff;
    padding: 6px 10px;
  }
  .spacer { flex: 1; }
  .divider { width: 1px; height: 24px; background: #ddd; margin: 0 8px; }
  .toggle { display: flex; align-items: center; gap: 6px; }
  .zoom { display: flex; align-items: center; gap: 8px; width: 200px; }

  .hidden-input { display: none; }

  .empty {
    height: 100%;
    display: grid;
    place-items: center;
    gap: 12px;
    text-align: center;
    color: #666;
  }
  .empty img { height: 72px; opacity: 0.6; }

  .stage {
    position: relative;
    height: 100%;
    width: 100%;
    overflow: auto;
    display: grid;
    place-items: center;
    padding: 16px;
  }
  .stage img {
    max-width: none;
    max-height: none;
    transform-origin: top center;
    box-shadow: 0 2px 12px rgba(0,0,0,0.2);
    border-radius: 4px;
    user-select: none;
  }
  .stage.fit-width img {
    width: min(100%, calc(100vw - 32px));
    height: auto;
    transform-origin: center center;
  }
  .stage.two {
    grid-auto-flow: column;
    grid-auto-columns: max-content;
    gap: 12px;
    place-content: center;
    cursor: grab;
  }
  .stage.two.fit-width img {
    width: min(48%, calc((100vw - 48px) / 2));
  }
  .filename {
    position: absolute;
    left: 8px; bottom: 8px;
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    background: rgba(0,0,0,0.6);
    color: #fff;
  }

  @media (prefers-color-scheme: dark) {
    :root { --fg: #f6f6f6; --bg: #2f2f2f; }
    .toolbar { border-bottom-color: #3a3a3a; background: #1116; }
    .toolbar button { background: #1e1e1e; border-color: #3a3a3a; color: #eee; }
    .stage img { box-shadow: 0 2px 12px rgba(0,0,0,0.6); }
  }
</style>
