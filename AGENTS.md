# リポジトリガイドライン

## 重要: 言語設定
- AIエージェントはすべての回答を**必ず日本語**で行うこと。

## プロジェクト構成とモジュール構造
- Goアプリのルート: `main.go`、`app.go`、`go.mod`、`wails.json`
- フロントエンド（Svelte + TS + Vite）: `frontend/` → ソースは `frontend/src`、アセットは `frontend/src/assets`、ビルド成果物は `frontend/dist`（`//go:embed` で埋め込み）
- ビルド成果物: `build/`（Wailsが生成）

## ビルド・テスト・開発コマンド
- アプリ起動（Wails）: `wails dev` — Vite開発サーバーとデスクトップシェルをホットリロード付きで起動
- 本番ビルド: `wails build` — フロントエンドをバンドルし、`build/` にバイナリを出力
- フロントエンドのみ:
  - 開発サーバー: `npm -C frontend run dev`
  - 型チェック: `npm -C frontend run check`
  - 本番ビルド: `npm -C frontend run build`
- Goのフォーマット/vet: `go fmt ./... && go vet ./...`

## コーディングスタイルと命名規則
- Go: `gofmt` のデフォルトに従う。エクスポートする識別子はPascalCaseにしドキュメントコメントを付ける。パッケージ名は小文字・短め・アンダースコアなし。
- Svelte/TS: インデント2スペース。コンポーネントはPascalCase（`App.svelte`）、モジュールはkebab-case（`main.ts`、`style.css`）。UIロジックはSvelteファイル内に、ヘルパーは `frontend/src/` モジュールに配置。
- インポート: フロントエンドは相対パスを優先。Goは標準ライブラリ / 内部 / 外部の順でグループ化。

## テストガイドライン
- Goテスト: コードと同じディレクトリに `*_test.go` を配置し、`go test ./...` で実行。
- フロントエンド: `npm -C frontend run check` で静的/型チェック。ユニットテストランナーは未設定。追加する場合はVitestを使用し、`src/**/*.spec.ts` として配置。

## コミットとプルリクエストのガイドライン
- コミット: 簡潔な命令形（例: `feat: 画像ビューア追加`、`fix: nil ctxのガード`）。変更はスコープを絞ること。
- PR: 明確な説明、関連Issue、確認手順を記載。UIの変更にはスクリーンショットや短い動画を添付。テスト済みプラットフォームを明記。

## セキュリティと設定
- Goバージョン: `go.mod` で定義。Node/Viteバージョンは `frontend/package.json` で固定。
- シークレットをコミットしないこと。設定は `wails.json` に記載し、認証情報やエンドポイントを追加する場合は実行時の環境変数を優先。
