package main

import (
	"image"
	"log"
	"os"

	"image/gif"

	"github.com/nfnt/resize"
	divoom "github.com/suapapa/go_divoom"
)

func main() {
	f, err := os.Open("pepe.gif")
	chk(err)

	g, err := gif.DecodeAll(f)
	chk(err)

	log.Println(g.Delay)

	imgs := make([]image.Image, len(g.Image)+10)
	for i := 0; i < len(g.Image); i++ {
		imgs[i] = resize.Resize(64, 64, g.Image[i], resize.Lanczos3)
	}
	for i := 0; i < 10; i++ {
		imgs[len(g.Image)+i] = imgs[len(g.Image)-1]
	}

	ds, err := divoom.FindDevice()
	chk(err)
	c := divoom.NewClient(ds[0])

	err = c.ResetSendingAnimationPicID()
	chk(err)
	err = c.SendAnimationImgs(1, g.Delay[0]*2, imgs)
	chk(err)
}

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
