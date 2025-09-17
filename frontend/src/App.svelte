<script lang="ts">
  import { ChooseAndOpenArchive, ListPages, GetPageDataURL } from '../wailsjs/go/main/App.js'

  let pages: string[] = []
  let index = 0
  let src: string | null = null
  let error: string | null = null

  async function loadCurrent() {
    if (pages.length === 0) { src = null; return }
    try {
      src = await GetPageDataURL(index)
      error = null
    } catch (e) {
      error = String(e)
    }
  }

  async function openZip() {
    try {
      pages = await ChooseAndOpenArchive()
      index = 0
      await loadCurrent()
    } catch (e) {
      error = String(e)
    }
  }

  async function next() {
    if (index + 1 < pages.length) {
      index += 1
      await loadCurrent()
    }
  }

  async function prev() {
    if (index > 0) {
      index -= 1
      await loadCurrent()
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'ArrowRight' || e.key === ' ') next()
    if (e.key === 'ArrowLeft'  || e.key === 'Backspace') prev()
  }
</script>

<main on:keydown={onKey} tabindex="0">
  <header class="toolbar">
    <button class="btn" on:click={openZip}>Open ZIP/CBZ</button>
    {#if pages.length > 0}
      <button class="btn" on:click={prev} disabled={index===0}>Prev</button>
      <span class="status">{index + 1} / {pages.length} — {pages[index]}</span>
      <button class="btn" on:click={next} disabled={index+1>=pages.length}>Next</button>
    {/if}
  </header>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if src}
    <div class="viewer">
      <img src={src} alt={pages[index]} />
    </div>
  {:else}
    <div class="placeholder">Open a ZIP/CBZ to start.</div>
  {/if}
</main>

<style>
  :global(html, body, #app) { height: 100%; }
  main { height: 100%; display: flex; flex-direction: column; }
  .toolbar {
    display: flex; gap: .5rem; align-items: center;
    padding: .5rem .75rem; border-bottom: 1px solid #223;
  }
  .btn { cursor: pointer; padding: .3rem .6rem; }
  .status { opacity: .8; }
  .viewer { flex: 1; display: grid; place-items: center; overflow: auto; background: #111; }
  img { max-width: 100%; max-height: 100%; object-fit: contain; }
  .placeholder { flex: 1; display: grid; place-items: center; color: #aaa; }
  .error { color: #f66; padding: .5rem .75rem; }
</style>
