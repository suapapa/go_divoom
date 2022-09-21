package main

import (
	"flag"
	"image"
	"log"
	"os"
	"strconv"
	"time"

	_ "image/jpeg"
	_ "image/png"

	"github.com/nfnt/resize"
	divoom "github.com/suapapa/go_divoom"
)

var (
	flagPrintDialInfo  bool
	flagFacesDemo      bool
	flagVisualizerDemo bool
	flagCloudDemo      bool
	flagAnimationDemo  bool
)

func main() {
	flag.BoolVar(&flagPrintDialInfo, "i", false, "print dials and diallists info")
	flag.BoolVar(&flagFacesDemo, "f", false, "faces demo")
	flag.BoolVar(&flagVisualizerDemo, "v", false, "visualizer demo")
	flag.BoolVar(&flagCloudDemo, "c", false, "cloud demo")
	flag.BoolVar(&flagAnimationDemo, "a", false, "animation demo")
	flag.Parse()

	if flagPrintDialInfo {
		// Dial Type
		dials, err := divoom.DialType()
		chk(err)
		log.Println("dials:")
		for _, d := range dials {
			dls, tot, err := divoom.DialList(d, 1) // divoom64 dial의 diallist는 모두 30개 미만이라 1 페이지 안에 들어옴.
			log.Printf("  %s(dl_tot=%d)\n", d, tot)
			chk(err)
			for _, dl := range dls {
				log.Printf("    %v\n", dl)
			}
		}
	}

	log.Println("===")

	devices, err := divoom.FindDevice()
	chk(err)
	if len(devices) < 1 {
		log.Fatal("no divoom device is found")
	}

	c := divoom.NewClient(devices[0])

	if flagFacesDemo {
		chanNumStr := flag.Arg(0)
		chanNum, err := strconv.Atoi(chanNumStr)
		chk(err)

		log.Printf("=== Faces chan: %d\n", chanNum)
		err = c.SelectFacesChannel(chanNum)
		chk(err)

		time.Sleep(3 * time.Second)
		log.Println("=== Faces chan")
		fID, err := c.GetSelectFaceID()
		chk(err)
		log.Println(fID)

		time.Sleep(3 * time.Second)
	}

	// 화면조정 화면만 나와서 임시로 막음
	/*
		time.Sleep(3 * time.Second)
		log.Println("=== Custom chan")
		err = c.SelectChannel(divoom.ChannelCustom)
		chk(err)
		c.CustomChannel(0)
		time.Sleep(1 * time.Second)
		c.CustomChannel(1)
		time.Sleep(1 * time.Second)
		c.CustomChannel(2)
		time.Sleep(1 * time.Second)
	*/

	// Visualizer
	if flagVisualizerDemo {
		time.Sleep(3 * time.Second)
		log.Println("=== Visualizer chan")
		for i := 0; i < 12; i++ {
			log.Printf("visualizer - %d\n", i)
			err = c.VisualizerChannel(i) // Custom chann, DIY Analog Clock
			chk(err)
			time.Sleep(3 * time.Second)
		}
	}

	// Cloud
	if flagCloudDemo {
		time.Sleep(3 * time.Second)
		log.Println("=== Cloud chan")
		c.CloudChannel(divoom.CloudChannelRecommendGallery)
	}

	//Animation
	if flagAnimationDemo {
		time.Sleep(3 * time.Second)
		log.Println("=== Animation")
		err := c.ResetSendingAnimationPicID()
		chk(err)

		var imgs []image.Image
		var delays []int
		for _, imgPath := range flag.Args() {
			f, err := os.Open(imgPath)
			chk(err)
			img, imgFmt, err := image.Decode(f)
			chk(err)
			img = resize.Resize(64, 64, img, resize.Lanczos3)
			imgs = append(imgs, img)
			delays = append(delays, 10*1000)
			log.Printf("imgPath: %s, imgFmt: %s\n", imgPath, imgFmt)
		}

		c.SendAnimationImgs(1, delays, imgs)

		picID, err := c.GetSendingAnimationPicID()
		chk(err)
		log.Printf("picID=%d\n", picID)
	}

	fID, err := c.GetSelectFaceID()
	chk(err)
	log.Println(fID)
}

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
