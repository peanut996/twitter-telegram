package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

func handleMessage(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil {
		return
	}
	if !validateJoinChannel(bot, update) {
		return
	}
	url := update.Message.Text
	if !isHTTPUrl(url) {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "请输入正确的视频地址"))
		return
	}
	if isM3U8VideoUrl(url) {
		sendM3U8VideoMessage(bot, update, url)
		return
	}
	if isStaticVideoUrl(url) {
		sendVideoMessage(bot, update, url)
		return
	}

	videoUrl, err := parseTwitterVideoUrl(url)
	if err != nil {
		log.Printf("get video url error: %v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "获取视频地址失败: "+err.Error()))
		return
	}
	sendVideoMessage(bot, update, videoUrl)
}

func sendM3U8VideoMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, url string) {
	err := m3u8Download(url)
	if err != nil {
		log.Printf("[bot] get m3u8 video url error: %v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "下载m3u8视频失败: "+err.Error()))
		return
	}
	defer cleanOutput()
	msg := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FilePath(Output))
	msg.ReplyToMessageID = update.Message.MessageID
	_, err = bot.Send(msg)
	if err != nil {
		msg := fmt.Sprintf("下载m3u8视频成功，但发送视频失败: %s\n\n%s", err.Error(), url)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
	}
}

func sendVideoMessage(bot *tgbotapi.BotAPI, update tgbotapi.Update, url string) {
	videoUrl := url
	msg := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FileURL(videoUrl))
	msg.ReplyToMessageID = update.Message.MessageID
	_, err := bot.Send(msg)
	if err != nil {
		msg := fmt.Sprintf("发送视频失败: %s\n\n%s", err.Error(), videoUrl)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
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
