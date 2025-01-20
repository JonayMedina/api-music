package services

import (
	"context"

	"github.com/JonayMedina/api-music/internal/cache/redis"
	"github.com/JonayMedina/api-music/internal/db/mongodb"
	"github.com/JonayMedina/api-music/internal/structs"
)

type SongService struct {
	repo       *mongodb.SongRepository
	cache      *redis.CacheService
	aggregator *MusicAggregator
}

func NewSongService(repo *mongodb.SongRepository, cache *redis.CacheService, aggregator *MusicAggregator) *SongService {
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

	// Buscar en la base de datos local
	songs, total, err := s.repo.SearchSongs(ctx, query, artist, album, page, limit)
	if err != nil {
		return nil, err
	}

	// Si no hay suficientes resultados, buscar en servicios externos
	if len(songs) < limit {
		externalSongs, err := s.aggregator.SearchAll(ctx, query, artist, album)
		if err == nil {
			// Guardar nuevos resultados en base de datos
			if err := s.repo.SaveSongs(ctx, externalSongs); err != nil {
				// logger.Error("Error saving songs", err)
			}

			// Actualizar búsqueda local
			songs, total, err = s.repo.SearchSongs(ctx, query, artist, album, page, limit)
			if err != nil {
				return nil, err
			}

			// Invalidar caché de búsquedas debido a nuevos resultados
			if err := s.cache.InvalidateSearches(ctx); err != nil {
				// logger.Error("Error invalidating cache", err)
			}
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
func (s *SongService) GetSongByID(ctx context.Context, id string) (*structs.Song, error) {
	// Intentar obtener del caché
	if cachedSong, err := s.cache.GetSong(ctx, id); err == nil && cachedSong != nil {
		return cachedSong, nil
	}

	// Obtener de la base de datos
	song, err := s.repo.GetSongByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Guardar en caché
	if song != nil {
		if err := s.cache.SetSong(ctx, song); err != nil {
			// logger.Error("Error caching song", err)
		}
	}

	return song, nil
}
