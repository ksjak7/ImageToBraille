package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"flag"
)

var threshold float64
var inverted *bool

func main() {
	imagePath := flag.String("path", "", "path to image to convert")
	outputPath := flag.String("output", "output.txt", "path for text file output")
	inputThreshold := flag.Float64("threshold", .44, "pixel color threshold")
	inverted = flag.Bool("invert", false, "inverts white and black")

	flag.Parse()
	threshold = max(min(*inputThreshold, 1), 0)

	//Output Flag values to user
	fmt.Println("Path: ", *imagePath)
	fmt.Println("Threshold: ", threshold)
	fmt.Println("Inverted: ", *inverted)
	fmt.Println("Output: ", *outputPath)

	image, err := getImageFromPath(*imagePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	//Sort data into BrailleData objects
	//Out of bounds returns RGBA (0, 0, 0, 0)
	brailleData := []BrailleData{}
	for i := image.Bounds().Min.Y; i < image.Bounds().Dy(); i += 4 {
		for j := image.Bounds().Min.X; j < image.Bounds().Dx(); j += 2 {
			brailleData = append(brailleData, BrailleData{
				colorToMono(image.At(j, i)),
				colorToMono(image.At(j, i+1)),
				colorToMono(image.At(j, i+2)),
				colorToMono(image.At(j+1, i)),
				colorToMono(image.At(j+1, i+1)),
				colorToMono(image.At(j+1, i+2)),
				colorToMono(image.At(j, i+3)),
				colorToMono(image.At(j+1, i+3)),
			})
		}
	}

	//build list of braille characters from the BrailleData
	brailleRunes := []rune{}
	for i, data := range brailleData {
		if i%(image.Bounds().Dx()/2) == 0 {
			brailleRunes = append(brailleRunes, '\n')
		}
		brailleRunes = append(brailleRunes, data.brailleDataToUnicode())
	}
	
	//Handle File Output
	fileOutput := string(brailleRunes)
	stream, err := os.Create(*outputPath); if err != nil {
		fmt.Println(err)
		return
	}

	defer stream.Close()
	stream.WriteString(fileOutput)
	stream.Sync()
}

type BrailleData struct {
	one   bool
	two   bool
	three bool
	four  bool
	five  bool
	six   bool
	seven bool
	eight bool
}

func getImageFromPath(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	if strings.HasSuffix(filePath, ".png") {
		return png.Decode(file)
	} else if strings.HasSuffix(filePath, ".jpg") || strings.HasSuffix(filePath, ".jpeg") {
		return jpeg.Decode(file)
	} else {
		return nil, errors.New("invalid file format")
	}
}

func colorToMono(color color.Color) bool {
	R, G, B, A := color.RGBA()
	// if (R != 65535 || G != 65535 || B != 65535) {
	// 	fmt.Println(R,G,B)
	// }
	luminance := (0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B))/float64(A)
	if *inverted {
		return luminance >= threshold
	} else {
		return luminance <= threshold
	}
}

func (b BrailleData) brailleDataToUnicode() rune {
	unicode := int(10240)
	if b.one {
		unicode += 1
	}
	if b.two {
		unicode += 2
	}
	if b.three {
		unicode += 4
	}
	if b.four {
		unicode += 8
	}
	if b.five {
		unicode += 16
	}
	if b.six {
		unicode += 32
	}
	if b.seven {
		unicode += 64
	}
	if b.eight {
		unicode += 128
	}

	return rune(unicode)
}

