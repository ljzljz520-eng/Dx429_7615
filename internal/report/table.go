package report

import (
	"fmt"
	"offlinebundle/internal/domain"
)

func PageRows(pages []domain.DocumentPage) []string {
	rows := make([]string, 0, len(pages))
	for _, p := range pages {
		rows = append(rows, fmt.Sprintf("%s | %s | %s", p.Status, p.URL, p.Path))
	}
	return rows
}
func AssetRows(assets []domain.AssetResource) []string {
	rows := make([]string, 0, len(assets))
	for _, a := range assets {
		rows = append(rows, fmt.Sprintf("%s | %s | %s", a.Kind, a.Status, a.URL))
	}
	return rows
}
func RenderTable(headers []string, rows []string) string {
	out := ""
	for _, h := range headers {
		out += h + "\t"
	}
	out += "\n"
	for _, r := range rows {
		out += r + "\n"
	}
	return out
}
