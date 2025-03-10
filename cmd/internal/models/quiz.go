package models

type Quiz struct {
	ID      int64    `json:"id"`
	Title   string   `json:"title"`
	Results []string `json:"results"`
}
