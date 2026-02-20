package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rsilvagit/go-work/internal/model"
)

// DiscordWriter sends jobs to a Discord channel via Webhook.
type DiscordWriter struct {
	webhookURL string
	client     *http.Client
}

func NewDiscordWriter(webhookURL string) *DiscordWriter {
	return &DiscordWriter{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}
}

func (dw *DiscordWriter) WriteJobs(jobs []model.Job) error {
	if len(jobs) == 0 {
		return dw.send("Nenhuma vaga encontrada.")
	}

	// Discord has a 2000 char limit per message. Split into chunks.
	var chunks []string
	var current strings.Builder
	header := fmt.Sprintf("**Encontradas %d vaga(s):**\n\n", len(jobs))
	current.WriteString(header)

	for i, j := range jobs {
		entry := formatDiscordJob(i+1, j)

		if current.Len()+len(entry) > 1900 {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		current.WriteString(entry)
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}

	for _, chunk := range chunks {
		if err := dw.send(chunk); err != nil {
			return err
		}
	}
	return nil
}

func formatDiscordJob(n int, j model.Job) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%d. %s**\n", n, j.Title)
	fmt.Fprintf(&b, "> Empresa: %s\n", j.Company)
	fmt.Fprintf(&b, "> Local: %s\n", j.Location)
	if j.WorkModel != "" {
		fmt.Fprintf(&b, "> Modelo: %s\n", j.WorkModel)
	}
	if j.JobType != "" {
		fmt.Fprintf(&b, "> Tipo: %s\n", j.JobType)
	}
	if j.URL != "" {
		fmt.Fprintf(&b, "> [Ver vaga](%s)\n", j.URL)
	}
	b.WriteString("\n")
	return b.String()
}

type discordPayload struct {
	Content string `json:"content"`
}

func (dw *DiscordWriter) send(text string) error {
	payload, err := json.Marshal(discordPayload{Content: text})
	if err != nil {
		return fmt.Errorf("discord: marshaling payload: %w", err)
	}

	resp, err := dw.client.Post(dw.webhookURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("discord: sending message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		return fmt.Errorf("discord: API error %d: %v", resp.StatusCode, result["message"])
	}

	return nil
}
