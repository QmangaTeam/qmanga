// Package metadatasource implements public catalogue APIs. No caller-supplied URL is fetched.
package metadatasource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"qmanga/internal/catalog/application"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type cacheItem struct {
	data    []byte
	expires time.Time
}
type Source struct {
	kind       string
	client     *http.Client
	mu         sync.Mutex
	cache      map[string]cacheItem
	cacheBytes int
	covers     map[string]string
	next       time.Time
	endpoint   string
}

func New(kind string) *Source {
	endpoint := "https://shikimori.one"
	if kind == "anilist" {
		endpoint = "https://graphql.anilist.co"
	}
	return &Source{kind: kind, endpoint: endpoint, client: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}, cache: map[string]cacheItem{}, covers: map[string]string{}}
}
func (s *Source) fetch(ctx context.Context, key, path string, body any) ([]byte, error) {
	s.mu.Lock()
	if item, ok := s.cache[key]; ok && time.Now().Before(item.expires) {
		s.mu.Unlock()
		return item.data, nil
	}
	if time.Now().Before(s.next) {
		s.mu.Unlock()
		return nil, application.ErrSourceLimited
	}
	delay := 2500 * time.Millisecond
	if s.kind == "shikimori" {
		delay = 600 * time.Millisecond
	}
	s.next = time.Now().Add(delay)
	s.mu.Unlock()
	method := "GET"
	var reader io.Reader
	if body != nil {
		method = "POST"
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	}
	request, err := http.NewRequestWithContext(ctx, method, s.endpoint+path, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Qmanga/1.0 (catalogue metadata import)")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, application.ErrSourceUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode == 429 {
		s.mu.Lock()
		delay := time.Minute
		if seconds, e := strconv.Atoi(response.Header.Get("Retry-After")); e == nil && seconds > 0 && seconds <= 3600 {
			delay = time.Duration(seconds) * time.Second
		}
		s.next = time.Now().Add(delay)
		s.mu.Unlock()
		return nil, application.ErrSourceLimited
	}
	if response.StatusCode != 200 {
		return nil, application.ErrSourceUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil || len(data) > 2*1024*1024 {
		return nil, application.ErrSourceUnavailable
	}
	s.mu.Lock()
	if len(s.cache) >= 256 || s.cacheBytes+len(data) > 8*1024*1024 {
		s.cacheBytes = 0
		s.cache = map[string]cacheItem{}
	}
	if previous, ok := s.cache[key]; ok {
		s.cacheBytes -= len(previous.data)
	}
	s.cacheBytes += len(data)
	s.cache[key] = cacheItem{data, time.Now().Add(10 * time.Minute)}
	s.mu.Unlock()
	return data, nil
}

var markup = regexp.MustCompile(`<[^>]*>|\[/?[a-zA-Z][^\]]*\]`)

func plain(value string, max int) string {
	value = html.UnescapeString(markup.ReplaceAllString(value, ""))
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > max {
		runes = runes[:max]
	}
	return string(runes)
}
func names(values []string, max int) []string {
	result := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.ReplaceAll(strings.ReplaceAll(plain(value, 200), "\n", " "), "\r", " ")
		if value != "" && !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
		if len(result) == max {
			break
		}
	}
	return result
}
func safeImage(value string, hosts ...string) string {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Port() != "" {
		return ""
	}
	for _, host := range hosts {
		if parsed.Hostname() == host {
			return value
		}
	}
	return ""
}
func initial(source, id string) application.ImportTitle {
	return application.ImportTitle{Source: source, ID: id, Aliases: []string{}, Authors: []string{}, Tags: []application.ExternalTag{}, GenreIDs: []string{}, CategoryIDs: []string{}, Candidates: []application.ImportCandidate{}}
}

type shikiTitle struct {
	ID                          int
	Name, Russian, Kind, Status string
	Japanese, English, Synonyms []string
	Description                 string
	AiredOn                     string `json:"aired_on"`
	Image                       struct{ Original string }
	Genres                      []struct {
		ID            int
		Name, Russian string
	}
	Publishers []struct{ Name string }
}

func (s *Source) shikiMap(item shikiTitle) application.ImportTitle {
	result := initial("shikimori", strconv.Itoa(item.ID))
	result.URL = "https://shikimori.one/mangas/" + result.ID
	result.Name = plain(item.Name, 200)
	result.OriginalName = first(item.Japanese)
	result.Aliases = names(append(append(append([]string{item.Russian}, item.English...), item.Synonyms...), item.Japanese...), 12)
	result.Description = plain(item.Description, 10000)
	if item.Kind == "manga" || item.Kind == "manhwa" || item.Kind == "manhua" {
		result.Kind = item.Kind
	}
	switch item.Status {
	case "released":
		result.OriginalStatus = "completed"
	case "ongoing":
		result.OriginalStatus = "ongoing"
	case "paused":
		result.OriginalStatus = "hiatus"
	case "discontinued":
		result.OriginalStatus = "cancelled"
	}
	if len(item.AiredOn) >= 4 {
		result.Year, _ = strconv.Atoi(item.AiredOn[:4])
	}
	image := item.Image.Original
	if strings.HasPrefix(image, "/") {
		image = "https://shikimori.one" + image
	}
	result.CoverURL = safeImage(image, "shikimori.one", "desu.shikimori.one", "moe.shikimori.one", "nyaa.shikimori.one")
	for _, tag := range item.Genres {
		result.Tags = append(result.Tags, application.ExternalTag{ID: strconv.Itoa(tag.ID), Name: plain(tag.Name, 160), Aliases: []string{plain(tag.Russian, 160)}})
	}
	s.rememberCover(result)
	return result
}
func first(values []string) string {
	if len(values) > 0 {
		return plain(values[0], 200)
	}
	return ""
}
func (s *Source) Search(ctx context.Context, query string) ([]application.ImportTitle, error) {
	if s.kind == "anilist" {
		return s.anilist(ctx, query, "")
	}
	data, err := s.fetch(ctx, "search:"+query, "/api/mangas?limit=10&search="+url.QueryEscape(query), nil)
	if err != nil {
		return nil, err
	}
	var items []shikiTitle
	if json.Unmarshal(data, &items) != nil {
		return nil, application.ErrSourceUnavailable
	}
	result := []application.ImportTitle{}
	for _, item := range items {
		if item.ID > 0 {
			result = append(result, s.shikiMap(item))
		}
	}
	return result, nil
}
func (s *Source) Detail(ctx context.Context, id string) (application.ImportTitle, error) {
	if s.kind == "anilist" {
		items, err := s.anilist(ctx, "", id)
		if err != nil {
			return application.ImportTitle{}, err
		}
		if len(items) != 1 || items[0].ID != id {
			return application.ImportTitle{}, application.ErrSourceUnavailable
		}
		return items[0], nil
	}
	data, err := s.fetch(ctx, "detail:"+id, "/api/mangas/"+id, nil)
	if err != nil {
		return application.ImportTitle{}, err
	}
	var item shikiTitle
	if json.Unmarshal(data, &item) != nil || strconv.Itoa(item.ID) != id {
		return application.ImportTitle{}, application.ErrSourceUnavailable
	}
	result := s.shikiMap(item)
	timer := time.NewTimer(650 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return application.ImportTitle{}, ctx.Err()
	case <-timer.C:
	}
	roles, err := s.fetch(ctx, "roles:"+id, "/api/mangas/"+id+"/roles", nil) // optional authors are fetched on a later selection when rate budget permits
	if err == nil {
		var values []struct {
			Roles  []string
			Person *struct{ Name string }
		}
		if json.Unmarshal(roles, &values) == nil {
			for _, role := range values {
				if role.Person == nil {
					continue
				}
				for _, kind := range role.Roles {
					if kind == "Story & Art" || kind == "Story" || kind == "Art" {
						result.Authors = append(result.Authors, plain(role.Person.Name, 160))
						break
					}
				}
			}
			result.Authors = names(result.Authors, 12)
		}
	}
	return result, nil
}

const mediaFields = `id title{romaji english native} synonyms description(asHtml:false) format status countryOfOrigin startDate{year} coverImage{large} genres tags{id name} staff(perPage:12){edges{role node{name{full}}}}`

type aniMedia struct {
	ID                                           int
	Title                                        struct{ Romaji, English, Native string }
	Synonyms                                     []string
	Description, Format, Status, CountryOfOrigin string
	StartDate                                    struct{ Year int }
	CoverImage                                   struct{ Large string }
	Genres                                       []string
	Tags                                         []struct {
		ID   int
		Name string
	}
	Staff struct {
		Edges []struct {
			Role string
			Node struct{ Name struct{ Full string } }
		}
	}
}

func (s *Source) anilist(ctx context.Context, query, id string) ([]application.ImportTitle, error) {
	variables := map[string]any{"search": query}
	gql := `query($search:String){Page(page:1,perPage:10){media(search:$search,type:MANGA){` + mediaFields + `}}}`
	if id != "" {
		numeric, _ := strconv.Atoi(id)
		variables = map[string]any{"id": numeric}
		gql = `query($id:Int){Media(id:$id,type:MANGA){` + mediaFields + `}}`
	}
	data, err := s.fetch(ctx, "ani:"+query+":"+id, "", map[string]any{"query": gql, "variables": variables})
	if err != nil {
		return nil, err
	}
	var response struct {
		Data struct {
			Page  struct{ Media []aniMedia }
			Media *aniMedia
		}
		Errors []json.RawMessage
	}
	if json.Unmarshal(data, &response) != nil || len(response.Errors) > 0 {
		return nil, application.ErrSourceUnavailable
	}
	items := response.Data.Page.Media
	if response.Data.Media != nil {
		items = []aniMedia{*response.Data.Media}
	}
	result := []application.ImportTitle{}
	for _, item := range items {
		if item.ID <= 0 {
			continue
		}
		value := initial("anilist", strconv.Itoa(item.ID))
		value.URL = "https://anilist.co/manga/" + value.ID
		value.Name = plain(item.Title.Romaji, 200)
		if value.Name == "" {
			value.Name = plain(item.Title.English, 200)
		}
		value.OriginalName = plain(item.Title.Native, 200)
		value.Aliases = names(append([]string{item.Title.English, item.Title.Romaji, item.Title.Native}, item.Synonyms...), 12)
		value.Description = plain(item.Description, 10000)
		if item.Format == "MANGA" || item.Format == "ONE_SHOT" {
			switch item.CountryOfOrigin {
			case "KR":
				value.Kind = "manhwa"
			case "CN":
				value.Kind = "manhua"
			default:
				value.Kind = "manga"
			}
		}
		switch item.Status {
		case "FINISHED":
			value.OriginalStatus = "completed"
		case "RELEASING":
			value.OriginalStatus = "ongoing"
		case "HIATUS":
			value.OriginalStatus = "hiatus"
		case "CANCELLED":
			value.OriginalStatus = "cancelled"
		}
		value.Year = item.StartDate.Year
		value.Country = item.CountryOfOrigin
		value.CoverURL = safeImage(item.CoverImage.Large, "s4.anilist.co")
		for _, genre := range item.Genres {
			value.Tags = append(value.Tags, application.ExternalTag{ID: "genre:" + genre, Name: plain(genre, 160), Aliases: []string{}})
		}
		for _, tag := range item.Tags {
			value.Tags = append(value.Tags, application.ExternalTag{ID: fmt.Sprint(tag.ID), Name: plain(tag.Name, 160), Aliases: []string{}})
		}
		for _, edge := range item.Staff.Edges {
			role := strings.ToLower(edge.Role)
			if strings.Contains(role, "story") || strings.Contains(role, "art") || strings.Contains(role, "original creator") {
				value.Authors = append(value.Authors, plain(edge.Node.Name.Full, 160))
			}
		}
		value.Authors = names(value.Authors, 12)
		s.rememberCover(value)
		result = append(result, value)
	}
	return result, nil
}

// OpenCover only follows a previously validated catalogue image URL; redirects are never followed.
func (s *Source) OpenCover(ctx context.Context, id string) (io.ReadCloser, error) {
	s.mu.Lock()
	cover := s.covers[id]
	s.mu.Unlock()
	if cover == "" {
		item, err := s.Detail(ctx, id)
		if err != nil {
			return nil, err
		}
		cover = item.CoverURL
	}
	if cover == "" {
		return nil, application.ErrSourceUnavailable
	}
	request, err := http.NewRequestWithContext(ctx, "GET", cover, nil)
	if err != nil {
		return nil, err
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, application.ErrSourceUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, application.ErrSourceUnavailable
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 25*1024*1024+1))
	if err != nil || len(data) > 25*1024*1024 {
		return nil, application.ErrSourceUnavailable
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *Source) rememberCover(item application.ImportTitle) {
	if item.CoverURL == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.covers) >= 256 {
		s.covers = map[string]string{}
	}
	s.covers[item.ID] = item.CoverURL
}
