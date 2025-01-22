package mysql

import (
	"context"
	"database/sql"
	"time"

	songsDb "github.com/JonayMedina/api-music-db/database/functions/songs"
	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
	"github.com/JonayMedina/api-music/internal/cache/redis"
	structs "github.com/JonayMedina/api-music/internal/structs"
)

type SongRepositoryMysql struct {
	db      *sql.DB
	cache   *redis.CacheService
	timeout time.Duration
}

func NewSongRepositoryMysql(db *sql.DB, cache *redis.CacheService) *SongRepositoryMysql {
	return &SongRepositoryMysql{db: db, cache: cache}
}

func (r *SongRepositoryMysql) SearchSongsFunction(ctx context.Context, query, artist, album string, page, limit int) ([]*dbStructs.Song, int64, error) {
	songs, total, err := songsDb.SearchSongs(query, artist, album, page, limit)
	if err != nil {
		return nil, 0, err
	}

	songPtrs := make([]*dbStructs.Song, len(songs))
	for i := range songs {
		songPtrs[i] = songs[i]
	}
	return songPtrs, total, nil
}

func (r *SongRepositoryMysql) SaveSongs(ctx context.Context, songs []*dbStructs.Song) error {
	newSongs := []*dbStructs.Song{}
	for _, song := range songs {
		song, err := songsDb.CreateSong(song)
		if err != nil {
			return err
		}
		newSongs = append(newSongs, song)
	}

	if err := r.cache.SetSearchResults(ctx, "", "", "", 0, 0, &structs.SearchResult{Songs: newSongs}); err != nil {
		return err
	}
	return nil
}

func (r *SongRepositoryMysql) GetSongByID(ctx context.Context, id int) (*dbStructs.Song, error) {
	return songsDb.GetSong(id)
}
