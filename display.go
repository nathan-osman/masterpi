package masterpi

import (
	"image"
	"image/draw"
	"io/ioutil"
	"os"

	"github.com/golang/freetype/truetype"
	"github.com/mdp/monochromeoled"
	"golang.org/x/exp/io/i2c"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	FontRegular = "regular"
	FontThin    = "thin"
)

type Display struct {
	oled  *monochromeoled.OLED
	img   *image.Gray
	fonts map[string]*truetype.Font
}

func loadFont(name string) *truetype.Font {
	r, err := FS.OpenFile(CTX, name, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	f, err := truetype.Parse(b)
	if err != nil {
		panic(err)
	}
	return f
}

func (d *Display) clearImage() {
	d.img = image.NewGray(image.Rect(0, 0, 128, 64))
	draw.Draw(d.img, d.img.Bounds(), image.Black, image.ZP, draw.Src)
}

func NewDisplay() (*Display, error) {
	o, err := monochromeoled.Open(
		&i2c.Devfs{Dev: "/dev/i2c-1"},
		0x3c,
		128,
		64,
	)
	if err != nil {
		return nil, err
	}
	d := &Display{
		oled: o,
		fonts: map[string]*truetype.Font{
			FontRegular: loadFont("fonts/roboto-regular.ttf"),
			FontThin:    loadFont("fonts/roboto-thin.ttf"),
		},
	}
	d.clearImage()
	return d, nil
}

func (d *Display) Clear() {
	d.clearImage()
}

func (d *Display) DrawText(text, fontName string, x, y, pointSize int) error {
	dr := &font.Drawer{
		Dst: d.img,
		Src: image.White,
		Face: truetype.NewFace(d.fonts[fontName], &truetype.Options{
			Size:    float64(pointSize),
			Hinting: font.HintingNone,
		}),
	}
	dr.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	dr.DrawString(text)
	return d.oled.SetImage(0, 0, d.img)
}

func (d *Display) Flip() error {
	return d.oled.Draw()
}

func (d *Display) Close() {
	d.oled.Close()
}
