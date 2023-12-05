package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Request struct {
	Url      string `json:"url"`
	VQuality string `json:"vQuality"`
}

type Response struct {
	Status string `json:"status"`
	Url    string `json:"url"`
}

func isStaticVideoUrl(url string) bool {
	return strings.Contains(url, "video.twimg.com") &&
		!strings.Contains(url, ".m3u8")
}

func isM3U8VideoUrl(url string) bool {
	return strings.Contains(url, "video.twimg.com") &&
		strings.Contains(url, ".m3u8")
}

func parseTwitterVideoUrl(url string) (string, error) {
	log.Printf("[parse] %s", url)

	originUrl := removeQueryString(url)
	videoUrl, err := getVideoUrl(originUrl)
	if err != nil {
		return "", err
	}
	return videoUrl, nil
}

func getVideoUrl(originUrl string) (string, error) {
	url := apiHost + apiJsonPath
	var request = Request{
		Url:      originUrl,
		VQuality: "max",
	}
	data, _ := json.Marshal(request)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Println("response body: ", string(body))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	if result.Status == "stream" {
		err = sendPreflightStreamRequest(result.Url)
		if err != nil {
			return "", err
		}
		return result.Url, nil
	} else if result.Status == "redirect" {
		return result.Url, nil
	}
	return "", errors.New(fmt.Sprintf("parse video failed, resp: %v", result))
}

func sendPreflightStreamRequest(url string) error {
	url = url + "&p=1"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println("response body: ", string(body))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}
	if result.Status != "continue" {
		return errors.New(fmt.Sprintf("send stream failed, resp: %v", result))
	}
	return nil
}
