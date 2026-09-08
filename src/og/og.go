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

var lineStartKinsoku = map[rune]bool{
	'、': true, '。': true, '，': true, '．': true,
	'！': true, '？': true, '!': true, '?': true,
	'）': true, '」': true, '』': true, '】': true, '〉': true, '》': true,
	'・': true, '：': true, '；': true, ':': true, ';': true,
}

var (
	boldFont    *opentype.Font
	regularFont *opentype.Font
	iconImg     image.Image
)

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func loadFont(path string) (*opentype.Font, error) {
	fontBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return opentype.Parse(fontBytes)
}

func init() {
	var err error
	boldFont, err = loadFont("./fonts/NotoSansJP-Bold.ttf")
	if err != nil {
		log.Fatalf("failed to load bold font: %v", err)
	}
	regularFont, err = loadFont("./fonts/NotoSansJP-Regular.ttf")
	if err != nil {
		log.Fatalf("failed to load regular font: %v", err)
	}

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
	displayW  = 1200
	displayH  = 630
	frame     = 40
	contentW  = displayW - frame*2
	contentH  = displayH - frame*2
	startX    = 105 // タイトルが始まる左端のX座標
	startY    = startX + 25
	sentenceW = contentW - (startX-frame)*2 + 20

	iconX     = startX - 20
	iconY     = (displayH - startX) + 25
	usernameX = iconX + 120
	usernameY = iconY - 50

	lineSpacing   = 0.95 // 行間倍率
	maxTitleLines = 2
	maxDescLines  = 3
)

func generateOG(w http.ResponseWriter, title string, description string) error {

	dc := gg.NewContext(displayW, displayH)

	// 背景と枠
	dc.SetHexColor(FRAME_COLOR)
	dc.Clear()
	dc.DrawRoundedRectangle(frame, frame, contentW, contentH, 40)
	dc.SetHexColor("#FFFFFF")
	dc.Fill()

	// アイコン
	dc.DrawImageAnchored(iconImg, iconX, iconY, 0.0, 1.0)

	regularFace, _ := opentype.NewFace(regularFont, &opentype.FaceOptions{Size: 40, DPI: 72, Hinting: font.HintingNone})
	defer regularFace.Close()

	boldFace, _ := opentype.NewFace(boldFont, &opentype.FaceOptions{Size: 60, DPI: 72, Hinting: font.HintingNone})
	defer boldFace.Close()

	// ユーザー名
	dc.SetFontFace(regularFace)
	dc.SetRGB255(80, 80, 80)
	dc.DrawStringAnchored(USER_NAME, usernameX, usernameY, 0.0, 0.25)

	// タイトルの描画行数と高さを計算
	dc.SetFontFace(boldFace)
	titleLines := wrapTextCJK(dc, title, sentenceW)
	titleLines = trim(titleLines, maxTitleLines)
	titleLineHeight := dc.FontHeight() * lineSpacing

	// 説明文の描画行数と高さを計算
	dc.SetFontFace(regularFace)
	descLines := wrapTextCJK(dc, description, sentenceW)
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

	currentY -= 10
	// 説明分を描画
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
