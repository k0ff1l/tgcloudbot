package telegram

import "mime/multipart"

const (
	tgApi  = "https://api.telegram.org/bot"
	chatId = "@testchatbotkostik"
)

var _ Bot = (*IBot)(nil)

// [https://core.telegram.org/bots/api#document]
// [https://core.telegram.org/bots/api#available-methods]

type Bot interface {
	// SendFile()
	// SendMessage()
	// EditMessage()
	// ...
}

type IBot struct {
}

func (b *IBot) SendMessage(msg string) error {
	// chat_id
	// text
	//

	return nil
}

func (b *IBot) SendAudio(msg string) error {
	// chat_id
	// audio (

	return nil
}

func (b *IBot) UploadFile(file *multipart.FileHeader) error {
	//
	//

	return nil
}
