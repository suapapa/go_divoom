package main

import (
	"flag"
	"log"
	"time"

	divoom "github.com/suapapa/go_divoom"
)

var (
	flagPrintDialInfo bool
)

func main() {
	flag.BoolVar(&flagPrintDialInfo, "i", false, "print dials and diallists info")
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

	time.Sleep(3 * time.Second)
	log.Println("=== Faces chan")
	err = c.SelectFacesChannel(12) // Faces chann, US Stock
	chk(err)

	time.Sleep(3 * time.Second)
	log.Println("=== Faces chan")
	fID, err := c.GetSelectFaceID() // {64, 100} Bitcoin=64, brightness=100
	chk(err)
	log.Println(fID)

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
	time.Sleep(3 * time.Second)
	log.Println("=== Visualizer chan")
	for i := 0; i < 12; i++ {
		log.Printf("visualizer - %d\n", i)
		err = c.VisualizerChannel(i) // Custom chann, DIY Analog Clock
		chk(err)
		time.Sleep(3 * time.Second)
	}

	// Cloud
	time.Sleep(3 * time.Second)
	log.Println("=== Cloud chan")
	c.CloudChannel(divoom.CloudChannelRecommendGallery)

}

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
