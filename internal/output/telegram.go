package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rsilvagit/go-work/internal/model"
)

// TelegramWriter sends jobs to a Telegram chat via the Bot API.
type TelegramWriter struct {
	token  string
	chatID string
	client *http.Client
}

func NewTelegramWriter(token, chatID string) *TelegramWriter {
	return &TelegramWriter{
		token:  token,
		chatID: chatID,
		client: &http.Client{},
	}
}

func (tw *TelegramWriter) WriteJobs(jobs []model.Job) error {
	if len(jobs) == 0 {
		return tw.send("Nenhuma vaga encontrada.")
	}

	// Telegram has a 4096 char limit per message. Split into chunks.
	var chunks []string
	var current strings.Builder
	header := fmt.Sprintf("*Encontradas %d vaga(s):*\n\n", len(jobs))
	current.WriteString(header)

	for i, j := range jobs {
		entry := formatJob(i+1, j)

		if current.Len()+len(entry) > 3800 {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		current.WriteString(entry)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	for _, chunk := range chunks {
		if err := tw.send(chunk); err != nil {
			return err
		}
	}
	return nil
}

func formatJob(n int, j model.Job) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*%d\\. %s*\n", n, escapeMarkdown(j.Title))
	fmt.Fprintf(&b, "Empresa: %s\n", escapeMarkdown(j.Company))
	fmt.Fprintf(&b, "Local: %s\n", escapeMarkdown(j.Location))
	fmt.Fprintf(&b, "Fonte: %s\n", escapeMarkdown(j.Source))
	if j.URL != "" {
		fmt.Fprintf(&b, "[Ver vaga](%s)\n", j.URL)
	}
	b.WriteString("\n")
	return b.String()
}

func escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
		">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	)
	return replacer.Replace(s)
}

func (tw *TelegramWriter) send(text string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tw.token)

	payload := map[string]string{
		"chat_id":    tw.chatID,
		"text":       text,
		"parse_mode": "MarkdownV2",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("telegram: marshaling payload: %w", err)
	}

	resp, err := tw.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: sending message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		return fmt.Errorf("telegram: API error %d: %v", resp.StatusCode, result["description"])
	}

	return nil
}
