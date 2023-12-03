package main

import (
	"net/url"
	"strings"
)

func isHTTPUrl(link string) bool {
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		return false
	}
	return true
}

func removeQueryString(u string) string {
	// 解析 URL
	parsedURL, err := url.Parse(u)
	if err != nil {
		return ""
	}

	// 清空 query string
	parsedURL.RawQuery = ""

	// 返回结果
	return parsedURL.String()
}
