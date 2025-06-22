package site

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SitemapGenerator 用于生成 sitemap 文件
type SitemapGenerator struct {
	OutputDir string // 文件保存目录
	Domain    string // 域名
	Entries   []string
}

// NewSitemapGenerator 创建一个新的 SitemapGenerator 实例
func NewSitemapGenerator(outputDir string, domain string) *SitemapGenerator {
	return &SitemapGenerator{
		OutputDir: outputDir,
		Domain:    domain,
		Entries:   []string{},
	}
}

// AddEntry 添加一行 sitemap 内容，包含 URL 和日期
func (sg *SitemapGenerator) AddEntry(url string, lastMod time.Time) {
	// 格式化为符合 Sitemap XML 的条目
	entry := fmt.Sprintf("<url><loc>%s</loc><lastmod>%s</lastmod></url>", url, lastMod.Format("2006-01-02"))
	sg.Entries = append(sg.Entries, entry)
}

// Save 将 sitemap 内容保存到文件
func (sg *SitemapGenerator) Save() error {
	if err := os.MkdirAll(sg.OutputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 根据条目数量选择保存方式
	if len(sg.Entries) > 50000 {
		return sg.saveMultiFile()
	}
	return sg.saveSingleFile()
}

// saveSingleFile 保存为单个文件
func (sg *SitemapGenerator) saveSingleFile() error {
	filePath := filepath.Join(sg.OutputDir, "sitemap.xml")
	content := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"
	for _, entry := range sg.Entries {
		content += entry + "\n"
	}
	content += "</urlset>"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

// saveMultiFile 保存为多文件，符合 Sitemap XML 规范
func (sg *SitemapGenerator) saveMultiFile() error {
	const maxEntriesPerFile = 50000
	var currentFileIndex int
	var currentFileContent string
	var currentEntryCount int
	var sitemapFiles []string

	for _, entry := range sg.Entries {
		currentFileContent += entry + "\n"
		currentEntryCount++

		// 如果当前文件的链接数量达到5万个，则保存当前文件并重置计数器
		if currentEntryCount >= maxEntriesPerFile {
			filePath := filepath.Join(sg.OutputDir, fmt.Sprintf("sitemap-%d.xml", currentFileIndex))
			content := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"
			content += currentFileContent
			content += "</urlset>"
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", filePath, err)
			}

			// 记录生成的 sitemap 文件
			sitemapFiles = append(sitemapFiles, filePath)

			// 重置内容和计数器，准备下一个文件
			currentFileContent = ""
			currentEntryCount = 0
			currentFileIndex++
		}
	}

	// 保存最后一个文件（可能不满5万个链接）
	if currentEntryCount > 0 { // 确保有剩余条目才保存
		filePath := filepath.Join(sg.OutputDir, fmt.Sprintf("sitemap-%d.xml", currentFileIndex))
		content := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"
		content += currentFileContent
		content += "</urlset>"
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}

		// 记录生成的 sitemap 文件
		sitemapFiles = append(sitemapFiles, filePath)
	}

	// 生成索引文件
	if err := sg.generateIndexFile(sitemapFiles); err != nil {
		return fmt.Errorf("failed to generate index file: %w", err)
	}

	return nil
}

// generateIndexFile 生成索引文件，记录所有生成的 sitemap 文件
func (sg *SitemapGenerator) generateIndexFile(sitemapFiles []string) error {
	indexFilePath := filepath.Join(sg.OutputDir, "sitemap-index.xml")

	// 确保目标目录已创建
	if err := os.MkdirAll(sg.OutputDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	content := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<sitemapindex xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"
	for _, file := range sitemapFiles {
		relativePath, err := filepath.Rel(sg.OutputDir, file)
		if err != nil {
			return fmt.Errorf("failed to get relative path for file %s: %w", file, err)
		}
		content += fmt.Sprintf("<sitemap><loc>%s/%s</loc><lastmod>%s</lastmod></sitemap>\n",
			sg.Domain, relativePath, time.Now().Format("2006-01-02"))
	}
	content += "</sitemapindex>"
	if err := os.WriteFile(indexFilePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write index file %s: %w", indexFilePath, err)
	}
	return nil
}
