package main

import (
	"errors"
	"fmt"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
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

	// TODO Lambda化して引数追加
	invalidateInput := makeInvalidate()

	slackPostErr1 := postSlack("キャッシュ削除開始します", true)
	if slackPostErr1 != nil {
		return "process not complete...", slackPostErr1
	}

	// キャッシュ削除実行
	_, invalidateError := doInvalidate(invalidateInput)

	invalidateResult := true
	if invalidateError != nil {
		invalidateResult = false
	}

	slackPostErr2 := postSlack("", invalidateResult)
	if slackPostErr2 != nil {
		return "process not complete...", slackPostErr2
	}

	return "process complete!", nil
}

// invalidationのInputを作成
func makeInvalidate(distributionId string, paths []*string) (cloudfront.CreateInvalidationInput, error) {

	// 削除の指定がない場合全件削除を追加
	if len(paths) == 0 {
		stab := "/*"
		paths = append(paths, &stab)
	}
	pathsLength := int64(len(paths))
	deleteItems := cloudfront.Paths{
		Items:    paths,
		Quantity: &pathsLength,
	}

	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	invalidationBatch := cloudfront.InvalidationBatch{
		CallerReference: &timeStamp,
		Paths:           &deleteItems,
	}

	invalidationInput := cloudfront.CreateInvalidationInput{
		DistributionId:    &distributionId,
		InvalidationBatch: &invalidationBatch,
	}

	return invalidationInput, nil
}

// キャッシュ削除処理
// 削除完了は待てないので、実行だけ。
func doInvalidate(input cloudfront.CreateInvalidationInput) (*cloudfront.CreateInvalidationOutput, error) {
	// Lambdaに実行権限を付与するのでClient情報はいらない
	client := cloudfront.CloudFront{
		Client: nil,
	}
	output, invalidateError := client.CreateInvalidation(&input)

	return output, invalidateError
}

// slackに連絡入れる
func postSlack(message string, status bool) error {
	webHookURL := os.Getenv("SLACK_WEB_HOOK_URL")
	iconUrl := os.Getenv("SLACK_ICON_URL")

	// messageが空文字の場合、固定のキャッシュ削除実行有無を指す文言をセットする
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
