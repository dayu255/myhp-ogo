package og

import (
	"encoding/base64"
	"image"
	"log"
	"net/http"
	"os"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var (
	boldFace    font.Face
	regularFace font.Face
	iconImg     image.Image
)

func loadFontFace(path string, points float64) (font.Face, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    points,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	return face, nil
}

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func init() {
	var err error

	// 1. フォントの事前読み込み
	boldFace, err = loadFontFace("./fonts/NotoSansJP-Bold.ttf", 60)
	if err != nil {
		log.Fatalf("failed to load bold font: %v", err)
	}
	regularFace, err = loadFontFace("./fonts/NotoSansJP-Regular.ttf", 40)
	if err != nil {
		log.Fatalf("failed to load regular font: %v", err)
	}

	// 2. 画像の事前読み込みとリサイズ
	rawImg, err := gg.LoadImage("./assets/icon.png")
	if err != nil {
		log.Fatalf("failed to load icon: %v", err)
	}
	iconImg = resizeImage(rawImg, 100, 100)
}

func decodeBase64(s string) string {
	if s == "" {
		return ""
	}
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return s
	}
	return string(decoded)
}

func Og(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	title := decodeBase64(params.Get("title"))
	description := decodeBase64(params.Get("description"))

	w.Header().Set("Content-Type", "image/png")
	if err := generateOG(w, title, description); err != nil {
		log.Printf(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func wrapTextCJK(dc *gg.Context, text string, maxWidth float64) string {
	var result string
	var currentLine string

	// 文字列を1文字(rune)ずつ処理
	for _, r := range text {
		testLine := currentLine + string(r)
		w, _ := dc.MeasureString(testLine)

		// 最大幅を超過した場合、手前で改行コードを挿入
		if w > maxWidth && currentLine != "" {
			result += currentLine + "\n"
			currentLine = string(r)
		} else {
			currentLine = testLine
		}
	}
	result += currentLine
	return result
}

func generateOG(w http.ResponseWriter, title string, description string) error {
	dc := gg.NewContext(1200, 630)

	// 背景と枠の描画
	dc.SetRGB255(232, 167, 172)
	dc.Clear()
	dc.DrawRoundedRectangle(50, 50, 1100, 530, 40)
	dc.SetRGB255(255, 255, 255)
	dc.Fill()

	// 事前ロード済みの画像を使用
	dc.DrawImageAnchored(iconImg, 125, 505, 0.5, 0.5)

	// 事前ロード済みのフォントを使用
	dc.SetFontFace(boldFace)
	dc.SetRGB255(51, 51, 51)
	wrappedTitle := wrapTextCJK(dc, title, 1000)
	dc.DrawStringWrapped(wrappedTitle, 100, 60, 0.0, 0.0, 1000, 1.0, gg.AlignLeft)

	dc.SetFontFace(regularFace)
	dc.SetRGB255(100, 100, 100)
	wrappedDesc := wrapTextCJK(dc, description, 1000)
	dc.DrawStringWrapped(wrappedDesc, 100, 260, 0.0, 0.0, 1000, 1.0, gg.AlignLeft)

	dc.SetRGB255(51, 51, 51)
	dc.DrawStringWrapped("dayu.jp", 200, 490, 0.0, 0.5, 1000, 1.5, gg.AlignLeft)

	return dc.EncodePNG(w)
}
