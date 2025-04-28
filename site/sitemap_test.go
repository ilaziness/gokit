package site

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestAddEntry 测试 AddEntry 方法是否正确添加条目
func TestAddEntry(t *testing.T) {
	sg := NewSitemapGenerator("test_output")
	sg.AddEntry("https://example.com", time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC))
	if len(sg.Entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(sg.Entries))
	}
	expectedEntry := "<url><loc>https://example.com</loc><lastmod>2023-10-01</lastmod></url>"
	if sg.Entries[0] != expectedEntry {
		t.Errorf("Expected entry: %s, got: %s", expectedEntry, sg.Entries[0])
	}
}

// TestSaveSingleFile 测试 Save 方法在条目数量小于等于 50000 时是否正确保存为单个文件
func TestSaveSingleFile(t *testing.T) {
	outputDir := "test_output_single"
	defer os.RemoveAll(outputDir) // 清理测试目录

	sg := NewSitemapGenerator(outputDir)
	sg.AddEntry("https://example.com", time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC))
	if err := sg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	filePath := filepath.Join(outputDir, "sitemap.xml")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expectedContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<url><loc>https://example.com</loc><lastmod>2023-10-01</lastmod></url>
</urlset>`
	if string(content) != expectedContent {
		t.Errorf("Expected content:\n%s\nGot:\n%s", expectedContent, string(content))
	}
}

// TestSaveMultiFile 测试 Save 方法在条目数量大于 50000 时是否正确保存为多个文件
func TestSaveMultiFile(t *testing.T) {
	outputDir := "test_output_multi"
	defer os.RemoveAll(outputDir) // 清理测试目录

	sg := NewSitemapGenerator(outputDir)
	for i := 0; i < 50001; i++ {
		sg.AddEntry(fmt.Sprintf("https://example.com/%d", i), time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC))
	}
	if err := sg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// 检查生成的文件数量
	files, err := filepath.Glob(filepath.Join(outputDir, "sitemap-*.xml"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	// 检查每个文件的内容
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("Failed to read file %s: %v", file, err)
		}
		expectedContent := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`
		if !strings.Contains(string(content), expectedContent) {
			t.Errorf("File %s does not contain expected content", file)
		}

		// 检查文件中的URL条目数量
		urlCount := strings.Count(string(content), "<url>")
		if filepath.Base(file) == "sitemap-0.xml" && urlCount != 50000 {
			t.Errorf("Expected 50000 URLs in sitemap-0.xml, got %d", urlCount)
		}
		if filepath.Base(file) == "sitemap-1.xml" && urlCount != 1 {
			t.Errorf("Expected 1 URL in sitemap-1.xml, got %d", urlCount)
		}
	}

	// 检查索引文件
	indexFilePath := filepath.Join(outputDir, "sitemap-index.xml")
	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		t.Fatalf("Failed to read index file: %v", err)
	}
	if !strings.Contains(string(content), "sitemap-0.xml") || !strings.Contains(string(content), "sitemap-1.xml") {
		t.Errorf("Index file does not contain expected entries")
	}
}

// TestGenerateIndexFile 测试 generateIndexFile 方法是否正确生成索引文件
func TestGenerateIndexFile(t *testing.T) {
	outputDir := "test_output_index"
	defer os.RemoveAll(outputDir) // 清理测试目录

	sg := NewSitemapGenerator(outputDir)
	sitemapFiles := []string{
		filepath.Join(outputDir, "sitemap-0.xml"),
		filepath.Join(outputDir, "sitemap-1.xml"),
	}
	if err := sg.generateIndexFile(sitemapFiles); err != nil {
		t.Fatalf("generateIndexFile failed: %v", err)
	}

	indexFilePath := filepath.Join(outputDir, "sitemap-index.xml")
	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		t.Fatalf("Failed to read index file: %v", err)
	}
	if !strings.Contains(string(content), "sitemap-0.xml") || !strings.Contains(string(content), "sitemap-1.xml") {
		t.Errorf("Index file does not contain expected entries")
	}
}
