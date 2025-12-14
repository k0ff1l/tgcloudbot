package telegram

// Chat represents a Telegram chat
// https://core.telegram.org/bots/api#chat
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

// User represents a Telegram user
// https://core.telegram.org/bots/api#user
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// Message represents a Telegram message
// https://core.telegram.org/bots/api#message
type Message struct {
	Photo     []PhotoSize `json:"photo,omitempty"`
	Text      string      `json:"text,omitempty"`
	Chat      Chat        `json:"chat"`
	MessageID int64       `json:"message_id"`
	Date      int64       `json:"date"`
	From      *User       `json:"from,omitempty"`
	Document  *Document   `json:"document,omitempty"`
	Audio     *Audio      `json:"audio,omitempty"`
	Video     *Video      `json:"video,omitempty"`
}

// Document represents a general file
// https://core.telegram.org/bots/api#document
type Document struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
}

// Audio represents an audio file
// https://core.telegram.org/bots/api#audio
type Audio struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Performer    string `json:"performer,omitempty"`
	Title        string `json:"title,omitempty"`
	FileName     string `json:"file_name,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	Duration     int    `json:"duration,omitempty"`
}

// Video represents a video file
// https://core.telegram.org/bots/api#video
type Video struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileName     string `json:"file_name,omitempty"`
	MimeType     string `json:"mime_type,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Duration     int    `json:"duration,omitempty"`
}

// PhotoSize represents one size of a photo
// https://core.telegram.org/bots/api#photosize
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FileSize     int64  `json:"file_size,omitempty"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

// File represents a file ready to be downloaded
// https://core.telegram.org/bots/api#file
type File struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	FilePath     string `json:"file_path,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
}

// Response represents a generic Telegram API response
// https://core.telegram.org/bots/api#making-requests
type Response struct {
	Description string              `json:"description,omitempty"`
	Result      *Message            `json:"result,omitempty"`
	Parameters  *ResponseParameters `json:"parameters,omitempty"`
	ErrorCode   int                 `json:"error_code,omitempty"`
	OK          bool                `json:"ok"`
}

// ResponseParameters represents parameters for error handling
type ResponseParameters struct {
	MigrateToChatID int64 `json:"migrate_to_chat_id,omitempty"`
	RetryAfter      int   `json:"retry_after,omitempty"`
}

// SendMessageRequest represents parameters for sendMessage method
type SendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"` // "HTML", "Markdown", "MarkdownV2"
}

// SendDocumentRequest represents parameters for sendDocument method
type SendDocumentRequest struct {
	ChatID              string `json:"chat_id"`
	Document            string `json:"document"` // file_id or URL
	Caption             string `json:"caption,omitempty"`
	ParseMode           string `json:"parse_mode,omitempty"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
}

// SendAudioRequest represents parameters for sendAudio method
type SendAudioRequest struct {
	ChatID              string `json:"chat_id"`
	Audio               string `json:"audio"`
	Caption             string `json:"caption,omitempty"`
	ParseMode           string `json:"parse_mode,omitempty"`
	Performer           string `json:"performer,omitempty"`
	Title               string `json:"title,omitempty"`
	Duration            int    `json:"duration,omitempty"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
}
