package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const baseURL = "http://localhost:9090/api/v1"

func main() {
	// 测试文章列表接口
	testArticleList()

	// 测试文章详情接口
	// testArticleGet()

	// 测试创建文章接口
	// testArticleCreate()

	// 测试更新文章接口
	// testArticleUpdate()

	// 测试删除文章接口
	// testArticleDelete()

	// 测试发布文章接口
	// testArticlePublish()

	// 测试归档文章接口
	// testArticleArchive()
}

func testArticleList() {
	url := fmt.Sprintf("%s/articles", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error testing article list: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Article List Response: %s\n", body)
}

func testArticleGet() {
	articleID := "test-article-id"
	url := fmt.Sprintf("%s/articles/%s", baseURL, articleID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error testing article get: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Article Get Response: %s\n", body)
}

func testArticleCreate() {
	url := fmt.Sprintf("%s/articles", baseURL)
	article := map[string]interface{}{
		"title":   "Test Article",
		"content": "This is a test article content",
		"summary": "This is a test article summary",
		"state":   "draft",
	}

	data, err := json.Marshal(article)
	if err != nil {
		fmt.Printf("Error marshaling article: %v\n", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error testing article create: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Article Create Response: %s\n", body)
}

func testArticleUpdate() {
	articleID := "test-article-id"
	url := fmt.Sprintf("%s/articles/%s", baseURL, articleID)
	article := map[string]interface{}{
		"title":   "Updated Test Article",
		"content": "Updated test article content",
		"summary": "Updated test article summary",
		"state":   "draft",
	}

	data, err := json.Marshal(article)
	if err != nil {
		fmt.Printf("Error marshaling article: %v\n", err)
		return
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(data))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error testing article update: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Article Update Response: %s\n", body)
}

func testArticleDelete() {
	articleID := "test-article-id"
	url := fmt.Sprintf("%s/articles/%s", baseURL, articleID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error testing article delete: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Article Delete Response: %s\n", body)
}

func testArticlePublish() {
	articleID := "test-article-id"
	url := fmt.Sprintf("%s/articles/%s/publish", baseURL, articleID)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error testing article publish: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Article Publish Response: %s\n", body)
}

func testArticleArchive() {
	articleID := "test-article-id"
	url := fmt.Sprintf("%s/articles/%s/archive", baseURL, articleID)

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error testing article archive: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Article Archive Response: %s\n", body)
}
