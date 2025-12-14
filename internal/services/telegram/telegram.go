package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	maxFileSize  = 50 * 1024 * 1024 // 50 MB - максимальный размер файла для Telegram
	sendDocument = "sendDocument"
	sendAudio    = "sendAudio"
	sendPhoto    = "sendPhoto"
	sendVideo    = "sendVideo"
)

var _ Client = (*Bot)(nil)

// Client interface defines methods for interacting with Telegram API
// https://core.telegram.org/bots/api#available-methods
type Client interface {
	// SendMessage sends a text message to a chat
	SendMessage(chatID, text string) (*Message, error)

	// SendDocument sends a document (file) to a chat
	SendDocument(chatID, filePath, caption string) (*Message, error)

	// SendAudio sends an audio file to a chat
	SendAudio(chatID, filePath, caption string) (*Message, error)

	// SendPhoto sends a photo to a chat
	SendPhoto(chatID, filePath, caption string) (*Message, error)

	// SendVideo sends a video file to a chat
	SendVideo(chatID, filePath, caption string) (*Message, error)

	// UploadFile uploads a file using multipart form data
	UploadFile(chatID string, file *multipart.FileHeader) (*Message, error)

	// GetFileInfo retrieves information about a file
	GetFileInfo(fileID string) (*File, error)
}

// Bot implements Client interface using HTTP requests to Telegram API
type Bot struct {
	client   *http.Client
	apiURL   string
	botToken string
}

// NewBot creates a new Telegram bot instance
func NewBot(apiURL, botToken string) *Bot {
	return &Bot{
		apiURL:   apiURL,
		botToken: botToken,
		client:   &http.Client{},
	}
}

func (b *Bot) SendMessage(chatID, text string) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	req := SendMessageRequest{ChatID: chatID, Text: text}

	return b.doJSONRequest("sendMessage", req)
}

func (b *Bot) SendDocument(chatID, filePath, caption string) (*Message, error) {
	return b.sendMultipartFile(sendDocument, "document", chatID, filePath, caption)
}

func (b *Bot) SendAudio(chatID, filePath, caption string) (*Message, error) {
	return b.sendMultipartFile(sendAudio, "audio", chatID, filePath, caption)
}

func (b *Bot) SendPhoto(chatID, filePath, caption string) (*Message, error) {
	return b.sendMultipartFile(sendPhoto, "photo", chatID, filePath, caption)
}

func (b *Bot) SendVideo(chatID, filePath, caption string) (*Message, error) {
	return b.sendMultipartFile(sendVideo, "video", chatID, filePath, caption)
}

func (b *Bot) UploadFile(chatID string, fileHeader *multipart.FileHeader) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if fileHeader == nil {
		return nil, errors.New("fileHeader is nil")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer safeClose(src)

	if fileHeader.Size > maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds limit %d", fileHeader.Size, maxFileSize)
	}

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	if err = writer.WriteField("chat_id", chatID); err != nil {
		return nil, fmt.Errorf("write chat_id failed: %w", err)
	}

	part, err := writer.CreateFormFile("document", fileHeader.Filename)
	if err != nil {
		return nil, fmt.Errorf("create form file failed: %w", err)
	}

	if _, err = io.Copy(part, src); err != nil {
		return nil, fmt.Errorf("copy file failed: %w", err)
	}

	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer failed: %w", err)
	}

	url := fmt.Sprintf("%s%s/sendDocument", b.apiURL, b.botToken)

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return b.doRequest(req)
}

func (b *Bot) GetFileInfo(fileID string) (*File, error) {
	if fileID == "" {
		return nil, errors.New("fileID cannot be empty")
	}

	url := fmt.Sprintf("%s%s/getFile?file_id=%s", b.apiURL, b.botToken, fileID)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}
	defer safeClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
	}

	var tgResp struct {
		Result      *File  `json:"result,omitempty"`
		Description string `json:"description,omitempty"`
		ErrorCode   int    `json:"error_code,omitempty"`
		OK          bool   `json:"ok"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram error %d: %s", tgResp.ErrorCode, tgResp.Description)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

func safeClose(c io.Closer) {
	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		log.Printf("failed to close resource: %v", err)
	}
}

func parseTelegramResponse(resp *http.Response) (*Message, error) {
	defer safeClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status %d: %s", resp.StatusCode, string(body))
	}

	var tgResp Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram error %d: %s", tgResp.ErrorCode, tgResp.Description)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

func (b *Bot) doRequest(req *http.Request) (*Message, error) {
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request failed: %w", err)
	}

	return parseTelegramResponse(resp)
}

func (b *Bot) doJSONRequest(method string, payload any) (*Message, error) {
	url := fmt.Sprintf("%s%s/%s", b.apiURL, b.botToken, method)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return b.doRequest(req)
}

func (b *Bot) sendMultipartFile(method, fieldName, chatID, filePath, caption string) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if filePath == "" {
		return nil, errors.New("filePath cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	defer safeClose(file)

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("get file info failed: %w", err)
	}

	if info.Size() > maxFileSize {
		return nil, fmt.Errorf("file size %d exceeds limit %d", info.Size(), maxFileSize)
	}

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)
	if err = writer.WriteField("chat_id", chatID); err != nil {
		return nil, fmt.Errorf("write chat_id failed: %w", err)
	}

	if caption != "" {
		if err = writer.WriteField("caption", caption); err != nil {
			return nil, fmt.Errorf("write caption failed: %w", err)
		}
	}

	part, err := writer.CreateFormFile(fieldName, filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file failed: %w", err)
	}

	if _, err = io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file failed: %w", err)
	}

	if err = writer.Close(); err != nil {
		return nil, fmt.Errorf("close writer failed: %w", err)
	}

	url := fmt.Sprintf("%s%s/%s", b.apiURL, b.botToken, method)

	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	return b.doRequest(req)
}
