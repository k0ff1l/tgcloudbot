package telegram

var _ Bot = (*IBot)(nil)

type Bot interface {
	// SendFile()
	// ...
}

type IBot struct {
}
