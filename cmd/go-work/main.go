package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rsilvagit/go-work/internal/filter"
	"github.com/rsilvagit/go-work/internal/model"
	"github.com/rsilvagit/go-work/internal/output"
	"github.com/rsilvagit/go-work/internal/scraper"
)

func loadEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func envOrFlag(flagVal, envKey string) string {
	if flagVal != "" {
		return flagVal
	}
	return os.Getenv(envKey)
}

func main() {
	loadEnv(".env")

	query := flag.String("q", "", "Termo de busca (ex: \"golang developer\")")
	location := flag.String("l", "", "Localização (ex: \"São Paulo\")")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout por scraper")
	telegramToken := flag.String("telegram-token", "", "Token do bot Telegram")
	telegramChatID := flag.String("telegram-chat-id", "", "Chat ID do Telegram")
	jobType := flag.String("tipo", "", "Tipo de vaga: full-time, part-time, estagio, freelance")
	workModel := flag.String("modelo", "", "Modelo: remoto, hibrido, presencial")
	level := flag.String("nivel", "", "Nível: junior, pleno, senior")
	region := flag.String("regiao", "", "Região/cidade para filtrar (ex: \"São Paulo\")")
	flag.Parse()

	if *query == "" {
		fmt.Fprintln(os.Stderr, "Erro: -q (query) é obrigatório")
		flag.Usage()
		os.Exit(1)
	}

	scrapers := scraper.Registry()
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	var (
		mu      sync.Mutex
		allJobs []model.Job
		wg      sync.WaitGroup
	)

	for _, s := range scrapers {
		wg.Add(1)
		go func(s scraper.Scraper) {
			defer wg.Done()
			fmt.Printf("Buscando em %s...\n", s.Name())
			jobs, err := s.Search(ctx, *query, *location)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Aviso: %s falhou: %v\n", s.Name(), err)
				return
			}
			mu.Lock()
			allJobs = append(allJobs, jobs...)
			mu.Unlock()
		}(s)
	}
	wg.Wait()

	// Deduplicate jobs by URL (or title+company).
	seen := make(map[string]bool)
	var uniqueJobs []model.Job
	for _, j := range allJobs {
		key := j.Key()
		if !seen[key] {
			seen[key] = true
			uniqueJobs = append(uniqueJobs, j)
		}
	}

	// Apply filters.
	uniqueJobs = filter.Apply(uniqueJobs, filter.Options{
		JobType:   *jobType,
		WorkModel: *workModel,
		Level:     *level,
		Region:    *region,
	})

	fmt.Println()
	writers := []output.ResultWriter{output.NewConsolePrinter()}

	tkn := envOrFlag(*telegramToken, "TELEGRAM_TOKEN")
	chatID := envOrFlag(*telegramChatID, "TELEGRAM_CHAT_ID")
	if tkn != "" && chatID != "" {
		writers = append(writers, output.NewTelegramWriter(tkn, chatID))
	}

	for _, w := range writers {
		if err := w.WriteJobs(uniqueJobs); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao exibir resultados: %v\n", err)
		}
	}

	fmt.Printf("\nTotal: %d vaga(s) encontrada(s).\n", len(uniqueJobs))
}
