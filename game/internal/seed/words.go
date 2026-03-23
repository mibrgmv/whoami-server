package seed

import (
	"bufio"
	"context"
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed */*.txt
var dataFS embed.FS

func Words(ctx context.Context, pool *pgxpool.Pool) error {
	var count int
	err := pool.QueryRow(ctx, "select count(*) from words").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check words count: %w", err)
	}

	if count > 0 {
		log.Printf("Words table already has %d entries, skipping seed", count)
		return nil
	}

	languages, err := dataFS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read seed directory: %w", err)
	}

	totalSeeded := 0

	for _, langDir := range languages {
		if !langDir.IsDir() {
			continue
		}

		language := langDir.Name()
		seeded, err := seedLanguage(ctx, pool, language)
		if err != nil {
			return fmt.Errorf("failed to seed %s: %w", language, err)
		}

		totalSeeded += seeded
	}

	log.Printf("Seeded %d words total", totalSeeded)
	return nil
}

func seedLanguage(ctx context.Context, pool *pgxpool.Pool, language string) (int, error) {
	solutionsPath := fmt.Sprintf("%s/solutions.txt", language)
	guessesPath := fmt.Sprintf("%s/guesses.txt", language)

	solutions, err := readWordList(solutionsPath)
	if err != nil {
		return 0, fmt.Errorf("failed to read solutions: %w", err)
	}

	guesses, err := readWordList(guessesPath)
	if err != nil {
		log.Printf("No guesses file for %s, using only solutions", language)
		guesses = nil
	}

	if len(solutions) == 0 {
		log.Printf("No solutions for language '%s', skipping", language)
		return 0, nil
	}

	count, err := insertWords(ctx, pool, language, solutions, true)
	if err != nil {
		return 0, err
	}

	if len(guesses) > 0 {
		guessCount, err := insertWords(ctx, pool, language, guesses, false)
		if err != nil {
			return 0, err
		}
		count += guessCount
	}

	log.Printf("Seeded %d words for language '%s' (%d solutions, %d guesses)",
		count, language, len(solutions), len(guesses))

	return count, nil
}

func readWordList(path string) ([]string, error) {
	data, err := dataFS.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var words []string
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}

	return words, scanner.Err()
}

func insertWords(ctx context.Context, pool *pgxpool.Pool, language string, words []string, isSolution bool) (int, error) {
	if len(words) == 0 {
		return 0, nil
	}

	const batchSize = 1000
	inserted := 0

	for i := 0; i < len(words); i += batchSize {
		end := i + batchSize
		if end > len(words) {
			end = len(words)
		}

		batch := words[i:end]
		query := buildInsertQuery(language, batch, isSolution)

		_, err := pool.Exec(ctx, query)
		if err != nil {
			return inserted, fmt.Errorf("failed to insert batch: %w", err)
		}

		inserted += len(batch)
	}

	return inserted, nil
}

func buildInsertQuery(language string, words []string, isSolution bool) string {
	var sb strings.Builder
	sb.WriteString("insert into words (word, language, is_solution) values ")

	for i, word := range words {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "('%s', '%s', %t)", word, language, isSolution)
	}

	sb.WriteString(" on conflict (word, language) do nothing")
	return sb.String()
}
