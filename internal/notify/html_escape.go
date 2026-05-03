package notify

import "html"

// EscapeTelegramHTML escapes text for Telegram Bot API HTML parse mode (entity text and attributes).
func EscapeTelegramHTML(s string) string {
	return html.EscapeString(s)
}
