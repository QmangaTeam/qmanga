package domain

import (
	"net/url"
	"strings"
	"unicode/utf8"
)

type ExternalLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}
type Publication struct {
	Year       int            `json:"year"`
	Publishers []string       `json:"publishers"`
	Source     string         `json:"source"`
	Country    string         `json:"country"`
	Formats    []string       `json:"formats"`
	Links      []ExternalLink `json:"links"`
}

func (p Publication) Valid() bool {
	if p.Year != 0 && (p.Year < 1000 || p.Year > 2200) || len(p.Publishers) > 12 || len(p.Formats) > 8 || len(p.Links) > 12 {
		return false
	}
	for _, value := range append(append([]string{p.Source, p.Country}, p.Publishers...), p.Formats...) {
		if utf8.RuneCountInString(value) > 160 || strings.ContainsAny(value, "\r\n") {
			return false
		}
	}
	for _, link := range p.Links {
		parsed, err := url.Parse(link.URL)
		if strings.TrimSpace(link.Label) == "" || utf8.RuneCountInString(link.Label) > 80 || strings.ContainsAny(link.Label, "\r\n") || len(link.URL) > 2000 || err != nil || parsed.Hostname() == "" || parsed.User != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
			return false
		}
	}
	return true
}
