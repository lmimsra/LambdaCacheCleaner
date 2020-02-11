package main

import (
	"errors"
	"fmt"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	// 後でLambda化
	_, _ = handle()
}

func handle() (string, error) {
	err := godotenv.Load("./.env")
	if err != nil {
		fmt.Println("[ERROR] env load error")
		return "process error!", errors.New("can not find env file")
	}

	slackPostErr1 := postSlack("キャッシュ削除開始します", true)
	if slackPostErr1 != nil {
		return "process not complete...", slackPostErr1
	}

	invalidateResult := doInvalidate()

	slackPostErr2 := postSlack("", invalidateResult)
	if slackPostErr2 != nil {
		return "process not complete...", slackPostErr2
	}

	return "process complete!", nil
}

// キャッシュ削除処理
// 削除完了は待てないので、実行だけ。
func doInvalidate() bool {
	return false
}

// slackに連絡入れる
func postSlack(message string, status bool) error {
	webHookURL := os.Getenv("SLACK_WEB_HOOK_URL")
	iconUrl := os.Getenv("SLACK_ICON_URL")
	if len(message) == 0 {
		message = map[bool]string{
			true:  "キャッシュ削除中\nそのうち削除されるのでしばしお待ちを",
			false: "キャッシュ削除失敗\nログを確認してください",
		}[status]
	}

	field := slack.Field{Title: "CloudFront", Value: message}
	attachment := slack.Attachment{}
	attachment.AddField(field)
	color := map[bool]string{true: "good", false: "danger"}[status]
	attachment.Color = &color
	payload := slack.Payload{
		Parse:       "",
		Username:    "aws-message",
		IconUrl:     iconUrl,
		IconEmoji:   "",
		Channel:     "",
		Text:        "",
		LinkNames:   "",
		Attachments: []slack.Attachment{attachment},
		UnfurlLinks: false,
		UnfurlMedia: false,
		Markdown:    false,
	}
	err := slack.Send(webHookURL, "", payload)
	if err != nil {
		fmt.Println("[ERROR] Slack post fail")
		return errors.New("SlackPostFail")
	}
	fmt.Println("[INFO] Slack post success")
	return nil
}
