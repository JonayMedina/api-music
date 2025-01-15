package itunes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/JonayMedina/api-music/internal/services"
	"github.com/JonayMedina/api-music/internal/structs"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type iTunesResponse struct {
	ResultCount int           `json:"resultCount"`
	Results     []iTunesTrack `json:"results"`
}

type iTunesTrack struct {
	TrackID         int     `json:"trackId"`
	TrackName       string  `json:"trackName"`
	ArtistName      string  `json:"artistName"`
	CollectionName  string  `json:"collectionName"`
	TrackTimeMillis int     `json:"trackTimeMillis"`
	ArtworkUrl100   string  `json:"artworkUrl100"`
	TrackPrice      float64 `json:"trackPrice"`
	Currency        string  `json:"currency"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) Name() string {
	return "itunes"
}

func (c *Client) Search(ctx context.Context, query, artist, album string) ([]structs.Song, error) {
	// Construir query
	searchQuery := query
	if artist != "" {
		searchQuery += " " + artist
	}
	if album != "" {
		searchQuery += " " + album
	}

	// Construir URL
	params := url.Values{}
	params.Add("term", searchQuery)
	params.Add("media", "music")
	params.Add("entity", "song")

	reqURL := fmt.Sprintf("%s/search?%s", c.baseURL, params.Encode())

	// Crear request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, services.NewServiceError("itunes", "failed to create request", err)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, services.NewServiceError("itunes", "failed to execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, services.NewServiceError("itunes", fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	// Decodificar respuesta
	var iTunesResp iTunesResponse
	if err := json.NewDecoder(resp.Body).Decode(&iTunesResp); err != nil {
		return nil, services.NewServiceError("itunes", "failed to decode response", err)
	}

	// Convertir resultados
	songs := make([]structs.Song, 0, len(iTunesResp.Results))
	for _, track := range iTunesResp.Results {
		songs = append(songs, structs.Song{
			ID:       strconv.Itoa(track.TrackID),
			Name:     track.TrackName,
			Artist:   track.ArtistName,
			Duration: formatDuration(track.TrackTimeMillis),
			Album:    track.CollectionName,
			Artwork:  track.ArtworkUrl100,
			Price:    formatPrice(track.TrackPrice, track.Currency),
			Origin:   "apple",
		})
	}

	return songs, nil
}

func formatDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, remainingSeconds)
}

func formatPrice(price float64, currency string) string {
	if price == 0 {
		return "Free"
	}
	return fmt.Sprintf("%s %.2f", currency, price)
}
