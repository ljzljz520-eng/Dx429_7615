package service

import (
	"offlinebundle/internal/domain"
	"sort"
)

type WorkQueue struct {
	items []string
	seen  map[string]bool
}

func NewQueue(seed string) *WorkQueue {
	q := &WorkQueue{items: []string{}, seen: map[string]bool{}}
	q.Add(seed)
	return q
}
func (q *WorkQueue) Add(value string) bool {
	if value == "" || q.seen[value] {
		return false
	}
	q.seen[value] = true
	q.items = append(q.items, value)
	return true
}
func (q *WorkQueue) Next() (string, bool) {
	if len(q.items) == 0 {
		return "", false
	}
	v := q.items[0]
	q.items = q.items[1:]
	return v, true
}
func (q *WorkQueue) Len() int { return len(q.items) }
func SortPages(pages []domain.DocumentPage) {
	sort.Slice(pages, func(i, j int) bool { return pages[i].URL < pages[j].URL })
}
