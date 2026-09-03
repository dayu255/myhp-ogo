package og

import (
	"image"
	"log"
	"net/http"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
)

func Og(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	title := params.Get("title")
	description := params.Get("description")
	w.Header().Set("Content-Type", "image/png")
	if err := generateOG(w, title, description); err != nil {
		log.Printf(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func generateOG(w http.ResponseWriter, title string, description string) error {
	dc := gg.NewContext(1200, 630)

	// 背景をピンクにする
	dc.SetRGB255(232, 167, 172)
	dc.Clear()

	// 外枠50
	dc.DrawRoundedRectangle(50, 50, 1100, 530, 40)
	dc.SetRGB255(255, 255, 255)
	dc.Fill()

	img, err := gg.LoadImage("./assets/icon.png")
	if err == nil {
		resizedImg := resizeImage(img, 100, 100)
		dc.DrawImageAnchored(resizedImg, 125, 505, 0.5, 0.5)
	}

	// 日本語フォントをサイズ60で読み込む
	if err := dc.LoadFontFace("./fonts/NotoSansJP-Bold.ttf", 60); err != nil {
		return err
	}
	// titleを描画
	dc.SetRGB255(51, 51, 51)
	dc.DrawStringWrapped(title, 100, 100, 0.0, 0.0, 1000, 1.5, gg.AlignLeft)

	if err := dc.LoadFontFace("./fonts/NotoSansJP-Regular.ttf", 40); err != nil {
		return err
	}
	// descriptionを描画
	dc.SetRGB255(100, 100, 100)
	dc.DrawStringWrapped(description, 100, 200, 0.0, 0.0, 1000, 1.5, gg.AlignLeft)

	dc.SetRGB255(51, 51, 51)
	dc.DrawStringWrapped("dayu.jp", 200, 505, 0.0, 0.5, 1000, 1.5, gg.AlignLeft)

	return dc.EncodePNG(w)
}
