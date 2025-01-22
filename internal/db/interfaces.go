package db

import (
	"context"
	"time"

	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
	"github.com/JonayMedina/api-music/internal/cache/redis"
)

// DatabaseProvider define la interfaz común para todos los proveedores de base de datos
type DatabaseProvider interface {
	SearchSongs(ctx context.Context, query, artist, album string, page, limit int) ([]*dbStructs.Song, int64, error)
	SaveSongs(ctx context.Context, songs []*dbStructs.Song) error
	GetSongByID(ctx context.Context, id interface{}) (*dbStructs.Song, error)
	Close() error
}

// DatabaseAggregator maneja múltiples proveedores de base de datos
type DatabaseAggregator struct {
	providers []DatabaseProvider
	cache     *redis.CacheService
}

// CacheProvider define la interfaz para el caché (opcional)
type CacheProvider interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Close() error
}
