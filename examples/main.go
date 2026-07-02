package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sazardev/go-pretty-pdf"
)

func main() {
	pdf, err := prettypdf.New(
		prettypdf.WithSourceDir("./examples/docs"),
		prettypdf.WithOutputFile("./examples/output/example-output.pdf"),
		prettypdf.WithTitle("go-pretty-pdf — Complete Example"),
		prettypdf.WithSubtitle("Every feature demonstrated with real MDX files"),
		prettypdf.WithAuthor("go-pretty-pdf Demo"),
		prettypdf.WithComponent("Callout", func(attrs map[string]string, inner string) string {
			level := attrs["title"]
			if level == "" {
				level = "info"
			}
			return fmt.Sprintf(
				`<div class="callout callout-%s"><strong>%s:</strong> %s</div>`,
				level, strings.ToUpper(level), inner,
			)
		}),
		prettypdf.WithComponent("Steps", func(attrs map[string]string, inner string) string {
			lines := strings.Split(strings.TrimSpace(inner), "\n")
			var items []string
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				items = append(items, fmt.Sprintf(
					`<div class="step-item"><span class="step-number">%d</span><span class="step-text">%s</span></div>`,
					i+1, line,
				))
			}
			return `<div class="steps-container">` + strings.Join(items, "\n") + `</div>`
		}),
		prettypdf.WithComponent("Card", func(attrs map[string]string, inner string) string {
			title := attrs["title"]
			icon := attrs["icon"]
			return fmt.Sprintf(
				`<div class="custom-card"><div class="card-header">%s <strong>%s</strong></div><div class="card-body">%s</div></div>`,
				icon, title, inner,
			)
		}),
	)
	if err != nil {
		log.Fatalf("Error creating PDF: %v", err)
	}

	fmt.Println("Building PDF from examples/docs/...")
	if err := pdf.Build(context.Background()); err != nil {
		log.Fatalf("Error building PDF: %v", err)
	}
	fmt.Println("✅ PDF generated: examples/output/example-output.pdf")
}
