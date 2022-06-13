package main

import (
	"log"
	"os"

	"image/gif"

	divoom "github.com/suapapa/go_divoom"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("gimme gif")
	}

	f, err := os.Open(os.Args[1])
	chk(err)

	g, err := gif.DecodeAll(f)
	chk(err)

	ds, err := divoom.FindDevice()
	chk(err)
	c := divoom.NewClient(ds[0])

	err = c.ResetSendingAnimationPicID()
	chk(err)
	err = c.SendAnimationGif(1, g)
	chk(err)
}

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
