package og

import (
	"encoding/base64"
	"image"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// icon.png path  from main.go
const ICON_PATH = "./assets/icon.png"
const USER_NAME = "dayu.jp"
const FRAME_COLOR = "#E8A7AC"

var (
	boldFace    font.Face
	regularFace font.Face
	iconImg     image.Image
)

var lineStartKinsoku = map[rune]bool{
	'、': true, '。': true, '，': true, '．': true,
	'！': true, '？': true, '!': true, '?': true,
	'）': true, '」': true, '』': true, '】': true, '〉': true, '》': true,
	'・': true, '：': true, '；': true, ':': true, ';': true,
}

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

	// フォントの読み込み
	boldFace, err = loadFontFace("./fonts/NotoSansJP-Bold.ttf", 60)
	if err != nil {
		log.Fatalf("failed to load bold font: %v", err)
	}
	regularFace, err = loadFontFace("./fonts/NotoSansJP-Regular.ttf", 40)
	if err != nil {
		log.Fatalf("failed to load regular font: %v", err)
	}

	// 画像の読み込みとリサイズ(100 * 100)
	rawImg, err := gg.LoadImage(ICON_PATH)
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
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func wrapTextCJK(dc *gg.Context, text string, maxWidth float64) []string {
	var lines []string

	// 既存の改行コードで分割
	rawParagraphs := strings.Split(text, "\n")

	for _, paragraph := range rawParagraphs {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		var currentLine []rune
		runes := []rune(paragraph)

		for i := 0; i < len(runes); i++ {
			r := runes[i]
			testLine := string(append(currentLine, r))
			w, _ := dc.MeasureString(testLine)

			if w > maxWidth && len(currentLine) > 0 {
				// 次の文字が行頭禁則文字の場合
				if lineStartKinsoku[r] {
					currentLine = append(currentLine, r)
					lines = append(lines, string(currentLine))
					currentLine = nil
					continue
				}

				lines = append(lines, string(currentLine))
				currentLine = []rune{r}
			} else {
				currentLine = append(currentLine, r)
			}
		}

		if len(currentLine) > 0 {
			lines = append(lines, string(currentLine))
		}
	}

	return lines
}

func trim(lines []string, maxLines int) []string {
	if len(lines) <= maxLines {
		return lines
	}

	lines = lines[:maxLines]
	lastLine := lines[maxLines-1]
	lastLine = lastLine[:len(lastLine)-2] + "..."
	lines[maxLines-1] = lastLine
	return lines
}

const (
	startX        = 125
	startY        = 135
	contentW      = 1000.0
	lineSpacing   = 1.00 // 行間倍率
	maxTitleLines = 2
	maxDescLines  = 3
)

func generateOG(w http.ResponseWriter, title string, description string) error {
	dc := gg.NewContext(1200, 630)

	// 背景と枠
	dc.SetHexColor(FRAME_COLOR)
	dc.Clear()
	dc.DrawRoundedRectangle(50, 50, 1100, 530, 40)
	dc.SetHexColor("#FFFFFF")
	dc.Fill()

	// アイコン
	dc.DrawImageAnchored(iconImg, startX, 505, 0.5, 0.5)

	// ユーザー名
	dc.SetFontFace(regularFace)
	dc.SetRGB255(80, 80, 80)
	dc.DrawStringAnchored(USER_NAME, startX+75, 505, 0.0, 0.25)

	// タイトルの描画行数と高さを計算
	dc.SetFontFace(boldFace)
	titleLines := wrapTextCJK(dc, title, contentW)
	titleLines = trim(titleLines, maxTitleLines)
	titleLineHeight := dc.FontHeight() * lineSpacing

	// 説明文の描画行数と高さを計算
	dc.SetFontFace(regularFace)
	descLines := wrapTextCJK(dc, description, contentW)
	descLines = trim(descLines, maxDescLines)
	descLineHeight := dc.FontHeight() * lineSpacing

	var currentX float64 = startX - 25
	var currentY float64 = startY

	// タイトルの描画
	dc.SetFontFace(boldFace)
	dc.SetRGB255(40, 40, 40)
	for _, line := range titleLines {
		dc.DrawStringAnchored(line, currentX, currentY, 0.0, 0.0)
		currentY += titleLineHeight
	}

	dc.SetFontFace(regularFace)
	dc.SetRGB255(110, 110, 110)
	for _, line := range descLines {
		dc.DrawStringAnchored(line, currentX, currentY, 0.0, 0.0)
		currentY += descLineHeight
	}

	log.Println(titleLines)
	log.Println(descLines)

	return dc.EncodePNG(w)
}
