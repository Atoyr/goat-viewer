# goat-viewer

コミック・画像アーカイブビューアーのデスクトップアプリ。
ZIP / CBZ ファイルを開いてページを1枚ずつ表示します。

## 機能

- **ファイルを開く** — ファイルダイアログから ZIP / CBZ アーカイブを選択して開く
- **ページ表示** — アーカイブ内の画像をページ単位で表示（背景色 #111 の中央配置）
- **ページナビゲーション**
  - ツールバーの `Prev` / `Next` ボタンでページ切り替え
  - キーボード操作: `→` or `Space` で次ページ、`←` or `Backspace` で前ページ
- **ページ情報表示** — 現在のページ番号 / 総ページ数、ファイル名をツールバーに表示
- **自然順ソート** — ファイル名を数値を考慮した自然順で並べる（例: `p2.jpg` < `p10.jpg`）
- **対応画像形式** — JPEG, PNG, GIF, WebP, BMP

## 対応アーカイブ形式

| 拡張子 | 説明 |
|--------|------|
| `.zip` | ZIP アーカイブ |
| `.cbz` | コミック用 ZIP アーカイブ |

## 開発

### 必要なもの

- [Go](https://go.dev/) 1.23+
- [Wails v2](https://wails.io/) v2.11.0+
- Node.js / npm

### コマンド

```bash
# ライブ開発（ホットリロードあり）
wails dev

# プロダクションビルド → build/ に出力
wails build

# フロントエンドのみ開発
npm -C frontend run dev

# TypeScript チェック
npm -C frontend run check

# Go フォーマット / 静的解析
go fmt ./... && go vet ./...

# テスト実行
go test ./...
```

## 技術スタック

| レイヤー | 技術 |
|----------|------|
| デスクトップフレームワーク | Wails v2 |
| バックエンド | Go 1.23 |
| フロントエンド | Svelte + TypeScript + Vite |

## アーキテクチャ

```
main.go          — Wails エントリポイント、frontend/dist を埋め込み
app.go           — Go バックエンド: アーカイブの開閉・ページ一覧・画像データ提供
frontend/
  src/
    App.svelte   — UI: ツールバー・画像ビューアー・キーボードナビ
    main.ts      — Svelte マウント
  wailsjs/       — Wails 自動生成バインディング
```

Go バックエンドは Wails バインディング経由でフロントエンドから呼び出されます。
画像は base64 データ URL としてフロントに渡されます。
