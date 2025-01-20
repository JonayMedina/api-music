package main

import (
	"github.com/JonayMedina/api-music/internal/cache/redis"
	"github.com/JonayMedina/api-music/internal/services"
)

type Services struct {
	SongService  *services.SongService
	CacheService *redis.CacheService
}

// func InitServices(cfg *config.Config) (*Services, error) {

// }
