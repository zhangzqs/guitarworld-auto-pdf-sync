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
		// 检测图片格式
		_, format, err := image.DecodeConfig(bytes.NewReader(imgData))
		if err != nil {
			return fmt.Errorf("failed to detect image format %d: %w", i+1, err)
		}

		var tmpFile string
		// 如果是 JPEG 或 PNG，直接使用原始数据，不重新编码
		if format == "jpeg" || format == "png" {
			ext := "jpg"
			if format == "png" {
				ext = "png"
			}
			tmpFile = filepath.Join(tempDir, fmt.Sprintf("page_%03d.%s", i+1, ext))
			if err := os.WriteFile(tmpFile, imgData, 0644); err != nil {
				return fmt.Errorf("failed to write image %d: %w", i+1, err)
			}
		} else {
			// 其他格式才需要解码并转换为 JPEG
			img, _, err := image.Decode(bytes.NewReader(imgData))
			if err != nil {
				return fmt.Errorf("failed to decode image %d: %w", i+1, err)
			}

			tmpFile = filepath.Join(tempDir, fmt.Sprintf("page_%03d.jpg", i+1))
			outFile, err := os.Create(tmpFile)
			if err != nil {
				return fmt.Errorf("failed to create temp file: %w", err)
			}

			// 使用适中质量编码为 JPEG
			err = jpeg.Encode(outFile, img, &jpeg.Options{Quality: 85})
			outFile.Close()
			if err != nil {
				return fmt.Errorf("failed to encode JPEG: %w", err)
			}
		}

		// 解码图片获取尺寸
		img, _, err := image.Decode(bytes.NewReader(imgData))
		if err != nil {
			return fmt.Errorf("failed to decode image %d for dimensions: %w", i+1, err)
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
