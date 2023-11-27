package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	apiHost     = "https://co.wuk.sh"
	apiJsonPath = "/api/json"

	telegramBotToken = ""

	debug = false

	channelName = ""
)

func init() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	debugMode := os.Getenv("DEBUG")
	channel := os.Getenv("TELEGRAM_CHANNEL_NAME")
	if token != "" {
		telegramBotToken = token
	}

	if telegramBotToken == "" {
		log.Fatal("telegram bot token is empty")
	}
	if debugMode == "true" {
		debug = true
	}
	if channel != "" {
		channelName = channel
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = debug
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if !validateJoinChannel(bot, update) {
				return
			}
			if isHTTPUrl(update) {
				handleMessage(update, bot)
			} else {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "请输入正确的视频地址"))
			}
		}
	}
}

func validateJoinChannel(b *tgbotapi.BotAPI, update tgbotapi.Update) bool {
	if channelName == "" {
		return true
	}
	ok := findMemberFromChat(b, channelName, update.Message.From.ID)
	channelUrl := "https://t.me/" + strings.ReplaceAll(channelName, "@", "")
	if !ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "请先加入频道")
		button1 := tgbotapi.InlineKeyboardButton{
			URL:  &channelUrl,
			Text: "频道(Channel)",
		}
		markup := tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{{button1}}}
		msg.ReplyMarkup = markup
		_, err := b.Send(msg)
		if err != nil {
			b.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "发送频道链接失败: "+err.Error()))
		}
		return false
	} else {
		return true
	}
}

func findMemberFromChat(b *tgbotapi.BotAPI, chatName string, userID int64) bool {
	findUserConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			SuperGroupUsername: chatName,
			UserID:             userID,
		},
	}
	member, err := b.GetChatMember(findUserConfig)
	if err != nil || member.Status == "left" || member.Status == "kicked" {
		log.Printf("[ShouldLimitUser] memeber should be limit. id: %d", userID)
		return false
	}
	return true
}

func isHTTPUrl(update tgbotapi.Update) bool {

	link := update.Message.Text
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		return false
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
