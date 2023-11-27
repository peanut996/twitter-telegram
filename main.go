package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	apiHost     = "https://co.wuk.sh"
	apiJsonPath = "/api/json"

	telegramBotToken = ""
)

func init() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token != "" {
		telegramBotToken = token
	}

	if telegramBotToken == "" {
		log.Fatal("telegram bot token is empty")
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if isHTTPUrl(update) {
				handleMessage(update, bot)
			} else {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "请输入正确的视频地址"))
			}
		}
	}
}

func isHTTPUrl(update tgbotapi.Update) bool {

	link := update.Message.Text
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		return false
	}

	// 检查域名或 IP 地址
	if _, err := net.LookupHost(link); err != nil {
		return false
	}

	// 检查端口号
	if strings.Contains(link, ":") {
		port, err := strconv.Atoi(link[strings.LastIndex(link, ":")+1:])
		if err != nil || port < 1 || port > 65535 {
			return false
		}
	}

	// 检查路径
	if strings.Contains(link, "/") {
		if !strings.HasPrefix(link[strings.Index(link, "/")+1:], "/") {
			return false
		}
	}

	return true
}

func handleMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	videoUrl, err := getVideoUrl(update.Message.Text)
	if err != nil {
		log.Printf("get video url error: %v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "获取视频地址失败: "+err.Error()))
		return
	}

	msg := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FileURL(videoUrl))
	msg.ReplyToMessageID = update.Message.MessageID
	_, err = bot.Send(msg)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "发送视频失败: "+err.Error()))
	}

}

func getVideoUrl(originUrl string) (string, error) {
	url := apiHost + apiJsonPath
	var request = struct {
		Url      string `json:"url"`
		VQuality string `json:"vQuality"`
	}{
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

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		var result struct {
			Status string `json:"status"`
			Url    string `json:"url"`
		}
		err = json.Unmarshal(body, &result)
		if err != nil {
			return "", err
		}

		return result.Url, nil
	} else {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}
}
