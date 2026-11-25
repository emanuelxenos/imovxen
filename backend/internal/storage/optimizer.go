package storage

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

const (
	MaxImageWidth = 1600
	JpegQuality   = 82 // Padrão da indústria: balanço perfeito entre nitidez e peso mínimo
)

// ProcessAndOptimizeImage lê uma imagem de qualquer resolução, redimensiona se for maior que 1600px e comprime em JPEG
func ProcessAndOptimizeImage(src io.Reader, dstPath string) error {
	img, format, err := image.Decode(src)
	if err != nil {
		return fmt.Errorf("erro ao decodificar imagem: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()

	var optimized image.Image = img
	if width > MaxImageWidth {
		// Redimensiona proporcionalmente mantendo aspecto e usando algoritmo Lanczos3 de alta fidelidade
		optimized = resize.Resize(MaxImageWidth, 0, img, resize.Lanczos3)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de destino: %w", err)
	}
	defer out.Close()

	if strings.ToLower(format) == "png" && width <= MaxImageWidth {
		return png.Encode(out, optimized)
	}

	// Comprime em JPEG otimizado
	options := &jpeg.Options{Quality: JpegQuality}
	return jpeg.Encode(out, optimized, options)
}
