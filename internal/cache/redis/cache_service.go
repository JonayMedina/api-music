package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
	"github.com/JonayMedina/api-music/internal/structs"
	"github.com/go-redis/redis"
)

const (
	// Prefijos de cache para diferentes tipos de datos
	searchPrefix = "search:"
	songPrefix   = "song:"

	// Tiempo de expiración del caché
	searchExpiration = 1 * time.Hour
	songExpiration   = 24 * time.Hour
)

type CacheService struct {
	client *RedisClient
}

func NewCacheService(client *RedisClient) *CacheService {
	return &CacheService{client: client}
}

func (cs *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	return cs.client.Get(ctx, key, dest)
}

func (cs *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cs.client.Set(ctx, key, value, expiration)
}

func (cs *CacheService) Close() error {
	return cs.client.Close()
}

// generateSearchKey genera una clave única para una búsqueda
func (cs *CacheService) generateSearchKey(query, artist, album string, page, limit int) string {
	// Crear un string único que representa la búsqueda
	searchStr := fmt.Sprintf("%s:%s:%s:%d:%d", query, artist, album, page, limit)

	// Generar hash para usar como clave
	hash := sha256.Sum256([]byte(searchStr))
	return searchPrefix + hex.EncodeToString(hash[:])
}

// GetSearchResults obtiene resultados de búsqueda del caché
func (cs *CacheService) GetSearchResults(ctx context.Context, query, artist, album string, page, limit int) (*structs.SearchResult, error) {
	key := cs.generateSearchKey(query, artist, album, page, limit)

	var result structs.SearchResult
	err := cs.client.Get(ctx, key, &result)
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SetSearchResults guarda resultados de búsqueda en el caché
func (cs *CacheService) SetSearchResults(ctx context.Context, query, artist, album string, page, limit int, results *structs.SearchResult) error {
	key := cs.generateSearchKey(query, artist, album, page, limit)
	return cs.client.Set(ctx, key, results, searchExpiration)
}

// GetSong obtiene una canción del caché
func (cs *CacheService) GetSong(ctx context.Context, id string) (*dbStructs.Song, error) {
	key := songPrefix + id

	var song *dbStructs.Song
	err := cs.client.Get(ctx, key, &song)
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return song, nil
}

// SetSong guarda una canción en el caché
func (cs *CacheService) SetSong(ctx context.Context, song *dbStructs.Song) error {
	key := songPrefix + strconv.Itoa(song.ID)
	return cs.client.Set(ctx, key, song, songExpiration)
}

// InvalidateSearches invalida todas las búsquedas en caché
func (cs *CacheService) InvalidateSearches(ctx context.Context) error {
	iter := cs.client.client.Scan(ctx, 0, searchPrefix+"*", 100).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())

		// Eliminar en lotes de 1000 claves
		if len(keys) >= 1000 {
			if err := cs.client.Del(ctx, keys...); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}

	// Eliminar las claves restantes
	if len(keys) > 0 {
		return cs.client.Del(ctx, keys...)
	}

	return iter.Err()
}
