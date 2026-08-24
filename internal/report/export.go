package report

import (
	"encoding/json"
	"sort"
	"strings"

	"offlinebundle/internal/domain"
)

type ExportRow struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Path   string `json:"path"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func Rows(pages []domain.DocumentPage, assets []domain.AssetResource) []ExportRow {
	rows := make([]ExportRow, 0, len(pages)+len(assets))
	for _, page := range pages {
		rows = append(rows, ExportRow{Kind: "page", URL: page.URL, Path: page.Path, Status: string(page.Status), Error: page.Error})
	}
	for _, asset := range assets {
		rows = append(rows, ExportRow{Kind: asset.Kind, URL: asset.URL, Path: asset.Path, Status: string(asset.Status), Error: asset.Error})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Kind == rows[j].Kind {
			return rows[i].URL < rows[j].URL
		}
		return rows[i].Kind < rows[j].Kind
	})
	return rows
}

func EncodeRows(pages []domain.DocumentPage, assets []domain.AssetResource) ([]byte, error) {
	return json.MarshalIndent(Rows(pages, assets), "", "  ")
}

func CSVRows(pages []domain.DocumentPage, assets []domain.AssetResource) string {
	lines := []string{"kind,url,path,status,error"}
	for _, row := range Rows(pages, assets) {
		lines = append(lines, strings.Join([]string{csvQuote(row.Kind), csvQuote(row.URL), csvQuote(row.Path), csvQuote(row.Status), csvQuote(row.Error)}, ","))
	}
	return strings.Join(lines, "\n") + "\n"
}

func csvQuote(value string) string {
	value = strings.ReplaceAll(value, "\"", "\"\"")
	return "\"" + value + "\""
}

func FilterRows(rows []ExportRow, status string) []ExportRow {
	if strings.TrimSpace(status) == "" {
		return append([]ExportRow(nil), rows...)
	}
	out := make([]ExportRow, 0)
	for _, row := range rows {
		if row.Status == status {
			out = append(out, row)
		}
	}
	return out
}

func StatusHistogram(rows []ExportRow) map[string]int {
	histogram := map[string]int{}
	for _, row := range rows {
		histogram[row.Status]++
	}
	return histogram
}

func NoticeLines(notices []domain.ExternalLinkNotice) []string {
	lines := make([]string, 0, len(notices))
	for _, notice := range notices {
		lines = append(lines, notice.Display())
	}
	sort.Strings(lines)
	return lines
}
