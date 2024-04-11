package main

import (
    "fmt"
    "image"
    "image/jpeg"
    "os"
    "github.com/nfnt/resize"
)



func CropAndResize() {
    inputFile, err := os.Open("input.jpg")
    if err != nil {
        fmt.Println("Error opening input file::", err)
        return
    }
    defer inputFile.Close()

    img, _, err := image.Decode(inputFile)
    if err != nil {
        fmt.Println("Error decoding image:", err)
        return
    }

  
    smallImg := resize.Thumbnail(100, 100, img, resize.Lanczos3)

    mediumImg := resize.Thumbnail(300, 300, img, resize.Lanczos3)

    largeImg := resize.Thumbnail(800, 800, img, resize.Lanczos3)

    outputDir := "./output/"

    if _, err := os.Stat(outputDir); os.IsNotExist(err) {
        os.Mkdir(outputDir, 0755)
    }

    saveImage(outputDir+"small.jpg", smallImg)
    saveImage(outputDir+"medium.jpg", mediumImg)
    saveImage(outputDir+"large.jpg", largeImg)

    fmt.Println("Images have been successfully cropped and resized!")
}

func saveImage(filename string, img image.Image) {
    outputFile, err := os.Create(filename)
    if err != nil {
        fmt.Println("Error creating output file::", err)
        return
    }
    defer outputFile.Close()

    err = jpeg.Encode(outputFile, img, nil)
    if err != nil {
        fmt.Println("Error saving image:", err)
        return
    }
}
