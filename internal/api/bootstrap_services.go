package api

import (
	"errors"
	"fmt"
	"log"

	"github.com/JonayMedina/api-music/internal/cache/redis"
	"github.com/JonayMedina/api-music/internal/config"
	"github.com/JonayMedina/api-music/internal/db"
	"github.com/JonayMedina/api-music/internal/db/mongodb"
	"github.com/JonayMedina/api-music/internal/db/mysql"
	"github.com/JonayMedina/api-music/internal/services"
	"github.com/JonayMedina/api-music/internal/services/chartlyrics"
	"github.com/JonayMedina/api-music/internal/services/itunes"
)

// Movemos la estructura Services aquí
type Services struct {
	SongService  *services.SongService
	CacheService *redis.CacheService
}

func InitServices(cfg *config.Config) (*Services, func(), error) {
	var (
		dbProviders []db.DatabaseProvider
		cache       *redis.CacheService
		cleanups    []func()
	)

	// Inicializar MySQL si está configurado
	if cfg.MySQLHost != "" {
		mysqlDB, err := mysql.InitDB(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("error iniciando MySQL: %v", err)
		}
		dbProviders = append(dbProviders, mysql.NewProvider(mysqlDB))
		cleanups = append(cleanups, func() { mysqlDB.Close() })
	}

	// Inicializar MongoDB si está configurado
	if cfg.MongoURI != "" {
		mongoClient, err := mongodb.InitDB(cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("error iniciando MongoDB: %v", err)
		}
		mongoRepo := mongodb.NewSongRepository(mongoClient)
		dbProviders = append(dbProviders, mongodb.NewProvider(mongoClient, mongoRepo))
		cleanups = append(cleanups, func() { mongoClient.Close() })
	}

	// Inicializar Redis si está configurado (opcional)
	if cfg.RedisURI != "" {
		redisClient, err := redis.NewRedisClient(cfg.RedisURI)
		if err != nil {
			log.Printf("Advertencia: Redis no disponible: %v", err)
		} else {
			cache = redis.NewCacheService(redisClient)
			cleanups = append(cleanups, func() { redisClient.Close() })
		}
	}

	// Verificar que al menos hay una base de datos configurada
	if len(dbProviders) == 0 {
		return nil, nil, errors.New("se requiere al menos una base de datos (MySQL o MongoDB)")
	}

	// Crear el repositorio con los proveedores disponibles
	repository, err := db.NewRepository(dbProviders, cache)
	if err != nil {
		return nil, nil, err
	}

	// Inicializar servicios de música
	itunesClient := itunes.NewClient(cfg.ITunesAPIURL)
	chartLyricsClient := chartlyrics.NewClient(cfg.ChartLyricsAPIURL)

	musicAggregator := services.NewMusicAggregator([]services.MusicProvider{
		itunesClient,
		chartLyricsClient,
	})

	cleanup := func() {
		for _, fn := range cleanups {
			fn()
		}
	}

	return &Services{
		SongService:  services.NewSongService(repository, cache, musicAggregator),
		CacheService: cache,
	}, cleanup, nil
}
