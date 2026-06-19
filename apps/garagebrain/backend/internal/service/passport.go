package service

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"

	"github.com/auto-brain/garagebrain/internal/model"
)

func GeneratePassport(car model.Car, records []model.ServiceRecord) ([]byte, error) {
	tmpl, err := template.ParseFiles("templates/passport.html")
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	totalCost := 0
	for _, r := range records {
		if r.Cost != nil {
			totalCost += *r.Cost
		}
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]any{
		"Car":      car,
		"Records":  records,
		"Date":     "2026-01-01",
		"TotalCost": totalCost,
	})
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	tmpHTML, err := os.CreateTemp("", "passport-*.html")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpHTML.Name())

	if _, err := tmpHTML.Write(buf.Bytes()); err != nil {
		tmpHTML.Close()
		return nil, err
	}
	tmpHTML.Close()

	tmpPDF := tmpHTML.Name() + ".pdf"
	defer os.Remove(tmpPDF)

	cmd := exec.Command("wkhtmltopdf",
		"--page-size", "A4",
		"--margin-top", "20mm",
		"--margin-bottom", "20mm",
		"--margin-left", "15mm",
		"--margin-right", "15mm",
		"--quiet",
		tmpHTML.Name(), tmpPDF,
	)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf: %w", err)
	}

	return os.ReadFile(tmpPDF)
}
