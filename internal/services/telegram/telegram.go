package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const (
	maxFileSize = 50 * 1024 * 1024 // 50 MB - максимальный размер файла для Telegram
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

// SendMessage sends a text message to the specified chat
// https://core.telegram.org/bots/api#sendmessage
func (b *Bot) SendMessage(chatID, text string) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if text == "" {
		return nil, errors.New("text cannot be empty")
	}

	req := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	url := fmt.Sprintf("%s%s/sendMessage", b.apiURL, b.botToken)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := b.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log but don't fail on close error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tgResp Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

// SendDocument sends a document (file) to the specified chat
// https://core.telegram.org/bots/api#senddocument
func (b *Bot) SendDocument(chatID, filePath, caption string) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if filePath == "" {
		return nil, errors.New("filePath cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log but don't fail on close error
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileInfo.Size(), maxFileSize)
	}

	// Create multipart form
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	// Add chat_id
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return nil, fmt.Errorf("failed to write chat_id: %w", err)
	}

	// Add caption if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return nil, fmt.Errorf("failed to write caption: %w", err)
		}
	}

	// Add file
	part, err := writer.CreateFormFile("document", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s%s/sendDocument", b.apiURL, b.botToken)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log but don't fail on close error
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tgResp Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

// SendAudio sends an audio file to the specified chat
// https://core.telegram.org/bots/api#sendaudio
func (b *Bot) SendAudio(chatID, filePath, caption string) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if filePath == "" {
		return nil, errors.New("filePath cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log but don't fail on close error
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileInfo.Size(), maxFileSize)
	}

	// Create multipart form
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	// Add chat_id
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return nil, fmt.Errorf("failed to write chat_id: %w", err)
	}

	// Add caption if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return nil, fmt.Errorf("failed to write caption: %w", err)
		}
	}

	// Add file
	part, err := writer.CreateFormFile("audio", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s%s/sendAudio", b.apiURL, b.botToken)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tgResp Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

// SendPhoto sends a photo to the specified chat
// https://core.telegram.org/bots/api#sendphoto
func (b *Bot) SendPhoto(chatID, filePath, caption string) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if filePath == "" {
		return nil, errors.New("filePath cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileInfo.Size(), maxFileSize)
	}

	// Create multipart form
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	// Add chat_id
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return nil, fmt.Errorf("failed to write chat_id: %w", err)
	}

	// Add caption if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return nil, fmt.Errorf("failed to write caption: %w", err)
		}
	}

	// Add file
	part, err := writer.CreateFormFile("photo", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s%s/sendPhoto", b.apiURL, b.botToken)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tgResp Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

// SendVideo sends a video file to the specified chat
// https://core.telegram.org/bots/api#sendvideo
func (b *Bot) SendVideo(chatID, filePath, caption string) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if filePath == "" {
		return nil, errors.New("filePath cannot be empty")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.Size() > maxFileSize {
		return nil, fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileInfo.Size(), maxFileSize)
	}

	// Create multipart form
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	// Add chat_id
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return nil, fmt.Errorf("failed to write chat_id: %w", err)
	}

	// Add caption if provided
	if caption != "" {
		if err := writer.WriteField("caption", caption); err != nil {
			return nil, fmt.Errorf("failed to write caption: %w", err)
		}
	}

	// Add file
	part, err := writer.CreateFormFile("video", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s%s/sendVideo", b.apiURL, b.botToken)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tgResp Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

// UploadFile uploads a file from multipart.FileHeader to Telegram
func (b *Bot) UploadFile(chatID string, fileHeader *multipart.FileHeader) (*Message, error) {
	if chatID == "" {
		return nil, errors.New("chatID cannot be empty")
	}

	if fileHeader == nil {
		return nil, errors.New("fileHeader is nil")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
		}
	}()

	if fileHeader.Size > maxFileSize {
		return nil, fmt.Errorf("file size (%d bytes) exceeds maximum allowed size (%d bytes)", fileHeader.Size, maxFileSize)
	}

	// Create multipart form
	var requestBody bytes.Buffer

	writer := multipart.NewWriter(&requestBody)

	// Add chat_id
	if err := writer.WriteField("chat_id", chatID); err != nil {
		return nil, fmt.Errorf("failed to write chat_id: %w", err)
	}

	// Add file
	part, err := writer.CreateFormFile("document", fileHeader.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s%s/sendDocument", b.apiURL, b.botToken)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tgResp Response
	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}

// GetFileInfo retrieves information about a file by its file_id
// https://core.telegram.org/bots/api#getfile
func (b *Bot) GetFileInfo(fileID string) (*File, error) {
	if fileID == "" {
		return nil, errors.New("fileID cannot be empty")
	}

	url := fmt.Sprintf("%s%s/getFile?file_id=%s", b.apiURL, b.botToken, fileID)

	resp, err := b.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status: %d, body: %s", resp.StatusCode, string(body))
	}

	var tgResp struct {
		Result      *File  `json:"result,omitempty"`
		Description string `json:"description,omitempty"`
		ErrorCode   int    `json:"error_code,omitempty"`
		OK          bool   `json:"ok"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tgResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !tgResp.OK {
		return nil, fmt.Errorf("telegram API error: %s (code: %d)", tgResp.Description, tgResp.ErrorCode)
	}

	if tgResp.Result == nil {
		return nil, errors.New("telegram API returned OK but result is nil")
	}

	return tgResp.Result, nil
}
