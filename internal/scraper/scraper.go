package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zhangzqs/guitarworld-auto-pdf-sync/internal/models"
)

const (
	APIBaseURL    = "https://user.guitarworld.com.cn/user/pu/my/pu_list"
	DetailBaseURL = "https://user.guitarworld.com.cn/user/pu/my"
	UserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36"
)

// Scraper 爬虫客户端
type Scraper struct {
	client    *http.Client
	cookies   string
	xsrfToken string
}

// New 创建新的 Scraper 实例
func New(cookies, xsrfToken string) *Scraper {
	return &Scraper{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		cookies:   cookies,
		xsrfToken: xsrfToken,
	}
}

// makeRequest 发起 HTTP 请求
func (s *Scraper) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,zh-CN;q=0.8,zh;q=0.7")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", "https://user.guitarworld.com.cn/user/pu/my")
	req.Header.Set("User-Agent", UserAgent)

	if s.xsrfToken != "" {
		req.Header.Set("X-XSRF-TOKEN", s.xsrfToken)
	}

	return s.client.Do(req)
}

// FetchSheetMusicPage 获取指定页的曲谱列表
func (s *Scraper) FetchSheetMusicPage(page int) (*models.APIResponse, error) {
	url := fmt.Sprintf("%s?page=%d", APIBaseURL, page)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Cookie", s.cookies)
	req.Header.Set("Referer", "https://user.guitarworld.com.cn/user/pu/my")
	req.Header.Set("User-Agent", UserAgent)
	if s.xsrfToken != "" {
		req.Header.Set("X-XSRF-TOKEN", s.xsrfToken)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page %d: %w", page, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp models.APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if apiResp.State != 1 {
		return nil, fmt.Errorf("API error: %s", apiResp.Message)
	}

	return &apiResp, nil
}

// FetchAllSheetMusic 获取所有曲谱
func (s *Scraper) FetchAllSheetMusic() ([]models.SheetMusicItem, error) {
	var allSheets []models.SheetMusicItem
	page := 1

	for {
		log.Printf("Fetching page %d...", page)

		apiResp, err := s.FetchSheetMusicPage(page)
		if err != nil {
			return nil, err
		}

		allSheets = append(allSheets, apiResp.Data.List...)

		log.Printf("Page %d: found %d items (total so far: %d)",
			page, len(apiResp.Data.List), len(allSheets))

		if len(apiResp.Data.List) == 0 ||
			(apiResp.Data.Pagination.LastPage > 0 &&
				apiResp.Data.Pagination.CurrentPage >= apiResp.Data.Pagination.LastPage) {
			log.Printf("Reached last page")
			break
		}

		page++
		time.Sleep(500 * time.Millisecond)
	}

	return allSheets, nil
}

// FetchSheetMusicImages 获取曲谱的图片 URL 列表
func (s *Scraper) FetchSheetMusicImages(id int) ([]string, error) {
	url := fmt.Sprintf("%s/%d", DetailBaseURL, id)

	resp, err := s.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch detail page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var imageURLs []string

	selectors := []string{
		"img.qupu-img",
		"img[src*='qupu']",
		"img[src*='pu_']",
		".qupu-content img",
		".pu-img img",
		"img[data-src]",
		"#qupu-container img",
		".score-image",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, sel *goquery.Selection) {
			if src, exists := sel.Attr("src"); exists && src != "" {
				imageURLs = append(imageURLs, src)
			}
			if dataSrc, exists := sel.Attr("data-src"); exists && dataSrc != "" {
				imageURLs = append(imageURLs, dataSrc)
			}
		})
		if len(imageURLs) > 0 {
			break
		}
	}

	if len(imageURLs) == 0 {
		doc.Find("img").Each(func(i int, sel *goquery.Selection) {
			src, exists := sel.Attr("src")
			if !exists {
				src, exists = sel.Attr("data-src")
			}
			if exists && src != "" {
				if !strings.Contains(src, "avatar") &&
					!strings.Contains(src, "logo") &&
					!strings.Contains(src, "icon") &&
					(strings.Contains(src, ".jpg") ||
						strings.Contains(src, ".jpeg") ||
						strings.Contains(src, ".png")) {
					imageURLs = append(imageURLs, src)
				}
			}
		})
	}

	uniqueURLs := make(map[string]bool)
	var result []string
	for _, url := range imageURLs {
		if !strings.HasPrefix(url, "http") {
			url = "https://www.guitarworld.com.cn" + url
		}
		if !uniqueURLs[url] {
			uniqueURLs[url] = true
			result = append(result, url)
		}
	}

	return result, nil
}

// DownloadImage 下载图片
func (s *Scraper) DownloadImage(url string) ([]byte, error) {
	resp, err := s.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return data, nil
}

// ExtractXSRFToken 从 cookies 中提取 XSRF-TOKEN
func ExtractXSRFToken(cookies string) string {
	re := regexp.MustCompile(`XSRF-TOKEN=([^;]+)`)
	matches := re.FindStringSubmatch(cookies)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
