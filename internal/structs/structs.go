package structs

import (
	"github.com/JonayMedina/api-music-db/database/structs"
)

// SearchRequest representa la estructura de búsqueda
type SearchRequest struct {
	Query  string `json:"query" form:"q"`
	Artist string `json:"artist" form:"artist"`
	Album  string `json:"album" form:"album"`
	Page   int    `json:"page" form:"page"`
	Limit  int    `json:"limit" form:"limit"`
}

// ErrorResponse representa la estructura de respuesta de error
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// APIResponse representa la estructura de respuesta general
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *MetaData   `json:"meta,omitempty"`
}

// MetaData representa la metadata de paginación
type MetaData struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
}

type SearchResult struct {
	Songs []*structs.Song `json:"songs"`
	Meta  MetaData        `json:"meta"`
}
