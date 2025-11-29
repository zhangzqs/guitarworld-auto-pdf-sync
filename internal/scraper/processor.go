package scraper

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/zhangzqs/guitarworld-auto-pdf-sync/internal/models"
	"github.com/zhangzqs/guitarworld-auto-pdf-sync/internal/pdf"
)

// Processor 曲谱处理器
type Processor struct {
	scraper   *Scraper
	outputDir string
}

// NewProcessor 创建新的 Processor 实例
func NewProcessor(scraper *Scraper, outputDir string) *Processor {
	return &Processor{
		scraper:   scraper,
		outputDir: outputDir,
	}
}

// ProcessSheet 处理单个曲谱
func (p *Processor) ProcessSheet(sheet models.SheetMusicItem, index, total int) models.ProcessResult {
	title := sheet.Name
	if sheet.SubTitle != "" {
		title = fmt.Sprintf("%s - %s", sheet.Name, sheet.SubTitle)
	}

	log.Printf("[%d/%d] Processing: %s (ID: %d) by %s", index, total, title, sheet.ID, sheet.CreatorName)

	outputPath := p.buildFilePath(sheet)

	// 确保创建者目录存在
	creatorDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(creatorDir, 0755); err != nil {
		return models.ProcessResult{Sheet: sheet, Success: false, Error: fmt.Errorf("create dir: %w", err)}
	}

	if _, err := os.Stat(outputPath); err == nil {
		log.Printf("[%d/%d] File already exists, skipping", index, total)
		return models.ProcessResult{Sheet: sheet, Success: true, Skipped: true}
	}

	// 获取曲谱图片URLs
	imageURLs, err := p.scraper.FetchSheetMusicImages(sheet.ID)
	if err != nil {
		return models.ProcessResult{Sheet: sheet, Success: false, Error: fmt.Errorf("fetch images: %w", err)}
	}

	if len(imageURLs) == 0 {
		return models.ProcessResult{Sheet: sheet, Success: false, Error: fmt.Errorf("no images found")}
	}

	log.Printf("[%d/%d] Found %d images, downloading...", index, total, len(imageURLs))

	// 下载所有图片到内存
	var imageDataList [][]byte
	for j, imgURL := range imageURLs {
		imgData, err := p.scraper.DownloadImage(imgURL)
		if err != nil {
			return models.ProcessResult{Sheet: sheet, Success: false, Error: fmt.Errorf("download image %d: %w", j+1, err)}
		}
		imageDataList = append(imageDataList, imgData)
		time.Sleep(200 * time.Millisecond)
	}

	// 创建PDF
	log.Printf("[%d/%d] Creating PDF with %d pages...", index, total, len(imageDataList))
	if err := pdf.CreateFromImageData(imageDataList, outputPath); err != nil {
		return models.ProcessResult{Sheet: sheet, Success: false, Error: fmt.Errorf("create PDF: %w", err)}
	}

	log.Printf("[%d/%d] Successfully created: %s", index, total, outputPath)
	return models.ProcessResult{Sheet: sheet, Success: true}
}

// sanitizeFileName 清理文件名中的非法字符
func sanitizeFileName(name string) string {
	// 只替换非法文件系统字符，保留空格
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	safe := reg.ReplaceAllString(name, "_")
	safe = strings.TrimSpace(safe)

	// 移除连续的下划线
	safe = regexp.MustCompile(`_+`).ReplaceAllString(safe, "_")
	safe = strings.Trim(safe, "_")

	return safe
}

// buildFilePath 构建文件路径
func (p *Processor) buildFilePath(sheet models.SheetMusicItem) string {
	creatorName := sanitizeFileName(sheet.CreatorName)
	if creatorName == "" {
		creatorName = "未知创建者"
	}

	// 获取曲谱类型
	categoryTxt := sheet.Qupu.CategoryTxt
	if categoryTxt == "" {
		categoryTxt = "吉他谱" // 默认值
	}

	// 构建文件名: [曲谱类型] 曲名 - 副标题
	fileName := fmt.Sprintf("[%s] %s", categoryTxt, sanitizeFileName(sheet.Name))
	if fileName == "" || sanitizeFileName(sheet.Name) == "" {
		fileName = fmt.Sprintf("[%s] sheet_%d", categoryTxt, sheet.ID)
	}

	if sheet.SubTitle != "" {
		subTitle := sanitizeFileName(sheet.SubTitle)
		fileName = fileName + " - " + subTitle
	}

	if len(fileName) > 200 {
		fileName = fileName[:200]
	}

	fileName = fileName + ".pdf"

	creatorDir := filepath.Join(p.outputDir, creatorName)

	return filepath.Join(creatorDir, fileName)
}
