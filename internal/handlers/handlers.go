package handlers

import (
	"net/http"
	"strconv"

	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
	"github.com/JonayMedina/api-music/internal/services"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	songService *services.SongService
}

func NewHandler(songService *services.SongService) *Handler {
	return &Handler{
		songService: songService,
	}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	songs := router.Group("/songs")
	{
		songs.GET("/search", h.SearchSongs)
		songs.GET("/:id", h.GetSong)
		songs.POST("/", h.SaveSong)
	}
}

func (h *Handler) SearchSongs(c *gin.Context) {
	query := c.Query("q")
	artist := c.Query("artist")
	album := c.Query("album")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.songService.SearchSongs(c.Request.Context(), query, artist, album, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) GetSong(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	song, err := h.songService.GetSongByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if song == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Canción no encontrada"})
		return
	}

	c.JSON(http.StatusOK, song)
}

func (h *Handler) SaveSong(c *gin.Context) {
	var song *dbStructs.Song

	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.songService.SaveSong(c.Request.Context(), song)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Canción guardada exitosamente"})
}
