package services

import (
	"context"
	"log"
	"strconv"

	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
	"github.com/JonayMedina/api-music/internal/cache/redis"
	"github.com/JonayMedina/api-music/internal/db"
	"github.com/JonayMedina/api-music/internal/structs"
)

type SongService struct {
	repo       *db.Repository
	cache      *redis.CacheService
	aggregator *MusicAggregator
}

func NewSongService(repo *db.Repository, cache *redis.CacheService, aggregator *MusicAggregator) *SongService {
	return &SongService{
		repo:       repo,
		cache:      cache,
		aggregator: aggregator,
	}
}

// SearchSongs busca canciones con soporte de caché
func (s *SongService) SearchSongs(ctx context.Context, query, artist, album string, page, limit int) (*structs.SearchResult, error) {
	// Intentar obtener del caché
	if cachedResult, err := s.cache.GetSearchResults(ctx, query, artist, album, page, limit); err == nil && cachedResult != nil {
		return cachedResult, nil
	}

	// Primero buscar en las bases de datos locales
	songs, total, err := s.repo.SearchSongs(ctx, query, artist, album, page, limit)
	if err != nil {
		return nil, err
	}

	if len(songs) < limit {
		externalSongs, err := s.aggregator.SearchAll(ctx, query, artist, album)
		if err == nil && len(externalSongs) > 0 {

			if err := s.repo.SaveSongs(ctx, externalSongs); err != nil {

				log.Printf("Error guardando canciones externas: %v", err)
			}
			songs = append(songs, externalSongs...)
		}
	}

	// Crear resultado
	result := &structs.SearchResult{
		Songs: songs,
		Meta: structs.MetaData{
			CurrentPage: page,
			PerPage:     limit,
			TotalPages:  (int(total) + limit - 1) / limit,
			TotalItems:  int(total),
		},
	}

	// Guardar en caché
	if err := s.cache.SetSearchResults(ctx, query, artist, album, page, limit, result); err != nil {
		// logger.Error("Error caching search results", err)
	}

	return result, nil
}

// GetSongByID obtiene una canción por ID con soporte de caché
func (s *SongService) GetSongByID(ctx context.Context, id int) (*dbStructs.Song, error) {
	// Intentar obtener del caché
	if cachedSong, err := s.cache.GetSong(ctx, strconv.Itoa(id)); err == nil && cachedSong != nil {
		return cachedSong, nil
	}

	// Obtener de la base de datos
	song, err := s.repo.GetSong(ctx, id)
	if err != nil {
		return nil, err
	}

	// Guardar en caché
	if song != nil {
		if err := s.cache.SetSong(ctx, song); err != nil {
			log.Println("Error caching song", err)
		}
	}

	return song, nil
}

func (s *SongService) SaveSong(ctx context.Context, song *dbStructs.Song) error {
	return s.repo.SaveSongs(ctx, []*dbStructs.Song{song})
}
