package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
	"github.com/JonayMedina/api-music/internal/cache/redis"
)

type Repository struct {
	aggregator *DatabaseAggregator
}

func NewRepository(providers []DatabaseProvider, cache *redis.CacheService) (*Repository, error) {
	if len(providers) == 0 {
		return nil, errors.New("al menos un proveedor de base de datos es requerido")
	}

	return &Repository{
		aggregator: &DatabaseAggregator{
			providers: providers,
			cache:     cache,
		},
	}, nil
}

func (r *Repository) SearchSongs(ctx context.Context, query, artist, album string, page, limit int) ([]*dbStructs.Song, int64, error) {
	// Intentar obtener del caché primero si está disponible
	if r.aggregator.cache != nil {
		var cachedResult struct {
			Songs []*dbStructs.Song
			Total int64
		}
		cacheKey := fmt.Sprintf("search:%s:%s:%s:%d:%d", query, artist, album, page, limit)
		if err := r.aggregator.cache.Get(ctx, cacheKey, &cachedResult); err == nil {
			return cachedResult.Songs, cachedResult.Total, nil
		}
	}

	// Búsqueda paralela en todas las bases de datos
	var (
		wg         sync.WaitGroup
		mu         sync.Mutex
		allSongs   []*dbStructs.Song
		totalCount int64
		errs       []error
	)

	for _, provider := range r.aggregator.providers {
		wg.Add(1)
		go func(p DatabaseProvider) {
			defer wg.Done()
			songs, count, err := p.SearchSongs(ctx, query, artist, album, page, limit)
			mu.Lock()
			if err != nil {
				errs = append(errs, err)
			} else {
				allSongs = append(allSongs, songs...)
				totalCount += count
			}
			mu.Unlock()
		}(provider)
	}

	wg.Wait()

	if len(errs) == len(r.aggregator.providers) {
		return nil, 0, fmt.Errorf("todos los proveedores fallaron: %v", errs)
	}

	// Guardar en caché si está disponible
	if r.aggregator.cache != nil {
		cacheKey := fmt.Sprintf("search:%s:%s:%s:%d:%d", query, artist, album, page, limit)
		r.aggregator.cache.Set(ctx, cacheKey, struct {
			Songs []*dbStructs.Song
			Total int64
		}{allSongs, totalCount}, time.Hour)
	}

	return allSongs, totalCount, nil
}

func (r *Repository) GetSong(ctx context.Context, id int) (*dbStructs.Song, error) {
	var (
		wg       sync.WaitGroup
		errChan  = make(chan error, len(r.aggregator.providers))
		songChan = make(chan *dbStructs.Song, 1)
	)

	for _, provider := range r.aggregator.providers {
		wg.Add(1)
		go func(p DatabaseProvider) {
			defer wg.Done()
			song, err := p.GetSongByID(ctx, id)
			if err != nil {
				errChan <- err
			} else if song != nil {
				songChan <- song
			}
		}(provider)
	}

	go func() {
		wg.Wait()
		close(errChan)
		close(songChan)
	}()

	select {
	case song := <-songChan:
		return song, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *Repository) SaveSongs(ctx context.Context, songs []*dbStructs.Song) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(r.aggregator.providers))

	for _, provider := range r.aggregator.providers {
		wg.Add(1)
		go func(p DatabaseProvider) {
			defer wg.Done()
			if err := p.SaveSongs(ctx, songs); err != nil {
				errChan <- err
			}
		}(provider)
	}

	wg.Wait()
	close(errChan)

	// Recolectar errores
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errores al guardar canciones: %v", errs)
	}

	return nil
}

func (r *Repository) Close() error {
	var errs []error

	for _, provider := range r.aggregator.providers {
		if err := provider.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.aggregator.cache != nil {
		if err := r.aggregator.cache.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errores al cerrar conexiones: %v", errs)
	}

	return nil
}
