package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ComicViewer メイン構造体
type ComicViewer struct {
	app         fyne.App
	window      fyne.Window
	imageWidget *canvas.Image
	statusLabel *widget.Label

	// データ管理
	zipReader    *zip.ReadCloser
	imageFiles   []string
	currentIndex int
}

// NewComicViewer 新しいビューワーを作成
func NewComicViewer() *ComicViewer {
	myApp := app.New()
	myApp.SetIcon(theme.DocumentIcon())

	window := myApp.NewWindow("漫画ビューワー")
	window.Resize(fyne.NewSize(800, 600))

	viewer := &ComicViewer{
		app:          myApp,
		window:       window,
		currentIndex: -1,
	}

	viewer.setupUI()
	return viewer
}

// setupUI UIを初期化
func (cv *ComicViewer) setupUI() {
	// 画像表示用ウィジェット
	cv.imageWidget = canvas.NewImageFromResource(theme.DocumentIcon())
	cv.imageWidget.FillMode = canvas.ImageFillContain
	cv.imageWidget.SetMinSize(fyne.NewSize(400, 300))

	// ステータスラベル
	cv.statusLabel = widget.NewLabel("ファイルを開いてください")

	// ボタン類
	openButton := widget.NewButton("Zipファイルを開く", cv.openZipFile)
	prevButton := widget.NewButton("前のページ", cv.previousPage)
	nextButton := widget.NewButton("次のページ", cv.nextPage)

	// ツールバー
	toolbar := container.NewHBox(
		openButton,
		widget.NewSeparator(),
		prevButton,
		nextButton,
	)

	// メインレイアウト
	content := container.NewBorder(
		toolbar,        // top
		cv.statusLabel, // bottom
		nil,            // left
		nil,            // right
		cv.imageWidget, // center
	)

	cv.window.SetContent(content)

	// キーボードショートカット
	cv.setupKeyboardShortcuts()
}

// setupKeyboardShortcuts キーボードショートカットを設定
func (cv *ComicViewer) setupKeyboardShortcuts() {
	cv.window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyLeft:
			cv.previousPage()
		case fyne.KeyRight:
			cv.nextPage()
		case fyne.KeyO:
			if key.Modifier == fyne.KeyModifierControl {
				cv.openZipFile()
			}
		}
	})
}

// openZipFile Zipファイルを開く
func (cv *ComicViewer) openZipFile() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser) {
		if reader == nil {
			return
		}
		defer reader.Close()

		cv.loadZipFile(reader.URI().Path())
	}, cv.window)

	// Zipファイルのフィルター
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".zip", ".cbz"}))
	fileDialog.Show()
}

// loadZipFile Zipファイルを読み込み
func (cv *ComicViewer) loadZipFile(path string) {
	// 既存のZipファイルを閉じる
	if cv.zipReader != nil {
		cv.zipReader.Close()
	}

	// 新しいZipファイルを開く
	reader, err := zip.OpenReader(path)
	if err != nil {
		dialog.ShowError(fmt.Errorf("Zipファイルを開けませんでした: %v", err), cv.window)
		return
	}

	cv.zipReader = reader
	cv.loadImageList()

	if len(cv.imageFiles) == 0 {
		dialog.ShowInformation("情報", "Zipファイル内に画像ファイルが見つかりませんでした", cv.window)
		return
	}

	// 最初の画像を表示
	cv.currentIndex = 0
	cv.showCurrentImage()
}

// loadImageList Zipファイル内の画像ファイル一覧を取得
func (cv *ComicViewer) loadImageList() {
	cv.imageFiles = []string{}

	// 対応する画像拡張子
	supportedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
	}

	for _, file := range cv.zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name))
		if supportedExts[ext] {
			cv.imageFiles = append(cv.imageFiles, file.Name)
		}
	}

	// ファイル名でソート
	sort.Strings(cv.imageFiles)
}

// showCurrentImage 現在のインデックスの画像を表示
func (cv *ComicViewer) showCurrentImage() {
	if cv.currentIndex < 0 || cv.currentIndex >= len(cv.imageFiles) {
		return
	}

	filename := cv.imageFiles[cv.currentIndex]

	// Zipファイル内から画像を読み込み
	imageData, err := cv.readImageFromZip(filename)
	if err != nil {
		dialog.ShowError(fmt.Errorf("画像を読み込めませんでした: %v", err), cv.window)
		return
	}

	// 画像をデコード
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		dialog.ShowError(fmt.Errorf("画像をデコードできませんでした: %v", err), cv.window)
		return
	}

	// キャンバスに設定
	cv.imageWidget.Image = img
	cv.imageWidget.Refresh()

	// ステータス更新
	cv.statusLabel.SetText(fmt.Sprintf("%d/%d - %s",
		cv.currentIndex+1, len(cv.imageFiles), filepath.Base(filename)))
}

// readImageFromZip Zipファイルから画像データを読み込み
func (cv *ComicViewer) readImageFromZip(filename string) ([]byte, error) {
	for _, file := range cv.zipReader.File {
		if file.Name == filename {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("ファイルが見つかりません: %s", filename)
}

// previousPage 前のページに移動
func (cv *ComicViewer) previousPage() {
	if cv.currentIndex > 0 {
		cv.currentIndex--
		cv.showCurrentImage()
	}
}

// nextPage 次のページに移動
func (cv *ComicViewer) nextPage() {
	if cv.currentIndex < len(cv.imageFiles)-1 {
		cv.currentIndex++
		cv.showCurrentImage()
	}
}

// Run アプリケーションを実行
func (cv *ComicViewer) Run() {
	cv.window.ShowAndRun()
}

// Close リソースを解放
func (cv *ComicViewer) Close() {
	if cv.zipReader != nil {
		cv.zipReader.Close()
	}
}

func main() {
	viewer := NewComicViewer()
	defer viewer.Close()
	viewer.Run()
}
