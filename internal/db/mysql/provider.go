package mysql

import (
	"context"
	"database/sql"

	songsDb "github.com/JonayMedina/api-music-db/database/functions/songs"
	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
)

type Provider struct {
	db *sql.DB
}

func NewProvider(db *sql.DB) *Provider {
	return &Provider{db: db}
}

func (p *Provider) SearchSongs(ctx context.Context, query, artist, album string, page, limit int) ([]*dbStructs.Song, int64, error) {
	return songsDb.SearchSongs(query, artist, album, page, limit)
}

func (p *Provider) SaveSongs(ctx context.Context, songs []*dbStructs.Song) error {
	for _, song := range songs {
		_, err := songsDb.CreateSong(song)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Provider) GetSongByID(ctx context.Context, id interface{}) (*dbStructs.Song, error) {
	return songsDb.GetSong(id.(int))
}

func (p *Provider) Close() error {
	return p.db.Close()
}
