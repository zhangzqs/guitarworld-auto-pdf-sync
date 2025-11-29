package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
)

// CreateFromImageData 从图片数据创建 PDF
func CreateFromImageData(imageDataList [][]byte, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "guitarworld-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	for i, imgData := range imageDataList {
		// 解码图片（自动识别格式：JPEG, PNG, GIF 等）
		img, _, err := image.Decode(bytes.NewReader(imgData))
		if err != nil {
			return fmt.Errorf("failed to decode image %d: %w", i+1, err)
		}

		// 统一转换为 JPEG 格式并保存到临时文件
		tmpFile := filepath.Join(tempDir, fmt.Sprintf("page_%03d.jpg", i+1))
		outFile, err := os.Create(tmpFile)
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}

		// 使用高质量编码为 JPEG
		err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 95})
		outFile.Close()
		if err != nil {
			return fmt.Errorf("failed to encode JPEG: %w", err)
		}

		// 获取图片尺寸
		imgBounds := img.Bounds()
		imgWidth := imgBounds.Dx()
		imgHeight := imgBounds.Dy()

		// 计算缩放比例以适应 A4 页面
		pageWidth := 210.0
		pageHeight := 297.0
		imgAspect := float64(imgWidth) / float64(imgHeight)
		pageAspect := pageWidth / pageHeight

		var pdfWidth, pdfHeight float64
		if imgAspect > pageAspect {
			pdfWidth = pageWidth
			pdfHeight = pageWidth / imgAspect
		} else {
			pdfHeight = pageHeight
			pdfWidth = pageHeight * imgAspect
		}

		pdf.AddPage()

		// 居中放置图片
		x := (pageWidth - pdfWidth) / 2
		y := (pageHeight - pdfHeight) / 2

		pdf.ImageOptions(tmpFile, x, y, pdfWidth, pdfHeight, false, gofpdf.ImageOptions{
			ReadDpi: true,
		}, 0, "")
	}

	return pdf.OutputFileAndClose(outputPath)
}
