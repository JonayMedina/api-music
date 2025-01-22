package chartlyrics

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
	"github.com/JonayMedina/api-music/internal/services"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type SearchLyricResponse struct {
	XMLName xml.Name `xml:"SearchLyricResponse"`
	Result  []struct {
		LyricID     int    `xml:"LyricId"`
		SongName    string `xml:"Song"`
		ArtistName  string `xml:"Artist"`
		AlbumName   string `xml:"Album"`
		TrackLyric  string `xml:"Lyric"`
		TrackLength string `xml:"SongRank"`
	} `xml:"SearchLyricResult"`
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
	return "chartlyrics"
}

func (c *Client) Search(ctx context.Context, query, artist, album string) ([]*dbStructs.Song, error) {
	// Construir URL
	params := url.Values{}
	if artist != "" {
		params.Add("artist", artist)
	}
	if query != "" {
		params.Add("song", query)
	}

	reqURL := fmt.Sprintf("%s/SearchLyric?%s", c.baseURL, params.Encode())

	// Crear request
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, services.NewServiceError("chartlyrics", "failed to create request", err)
	}

	// Ejecutar request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, services.NewServiceError("chartlyrics", "failed to execute request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, services.NewServiceError("chartlyrics", fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	// Decodificar respuesta XML
	var searchResp SearchLyricResponse
	if err := xml.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, services.NewServiceError("chartlyrics", "failed to decode response", err)
	}

	// Convertir resultados
	songs := make([]*dbStructs.Song, 0, len(searchResp.Result))
	for _, result := range searchResp.Result {
		songs = append(songs, &dbStructs.Song{
			Title:       result.SongName,
			Album:       result.AlbumName,
			Genre:       "",
			ReleaseDate: "",
			CoverImage:  "",
			Origin:      "chartlyrics",
		})
	}

	return songs, nil
}
