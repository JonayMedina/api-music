package services

import (
	"context"
	"sort"
	"sync"

	"github.com/JonayMedina/api-music-db/database/structs"
)

type MusicAggregator struct {
	providers []MusicProvider
}

func NewMusicAggregator(providers []MusicProvider) *MusicAggregator {
	return &MusicAggregator{
		providers: providers,
	}
}

func (ma *MusicAggregator) SearchAll(ctx context.Context, query, artist, album string) ([]*structs.Song, error) {
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		allSongs  []*structs.Song
		errChan   = make(chan error, len(ma.providers))
		songsChan = make(chan []*structs.Song, len(ma.providers))
	)

	// Lanzar búsquedas en paralelo
	for _, provider := range ma.providers {
		wg.Add(1)
		go func(p MusicProvider) {
			defer wg.Done()

			songs, err := p.Search(ctx, query, artist, album)
			if err != nil {
				errChan <- err
				return
			}

			songsChan <- songs
		}(provider)
	}

	// Esperar a que todas las búsquedas terminen
	go func() {
		wg.Wait()
		close(songsChan)
		close(errChan)
	}()

	// Recolectar resultados
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	for songs := range songsChan {
		mu.Lock()
		allSongs = append(allSongs, songs...)
		mu.Unlock()
	}

	// Ordenar resultados por nombre
	sort.Slice(allSongs, func(i, j int) bool {
		return allSongs[i].Title < allSongs[j].Title
	})

	return allSongs, nil
}
