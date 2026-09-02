package main

import (
    "fmt"
    "github.com/z-wentao/PhotoShare/models"
)

func main() {
    gs := models.GalleryService{}
    fmt.Println(gs.Images(6))
}
