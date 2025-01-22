package mongodb

import (
	"context"

	dbStructs "github.com/JonayMedina/api-music-db/database/structs"
)

type Provider struct {
	client     *MongoClient
	repository *SongRepository
}

func NewProvider(client *MongoClient, repository *SongRepository) *Provider {
	return &Provider{client: client, repository: repository}
}

func (p *Provider) SearchSongs(ctx context.Context, query, artist, album string, page, limit int) ([]*dbStructs.Song, int64, error) {
	songs, total, err := p.repository.SearchSongs(ctx, query, artist, album, page, limit)
	if err != nil {
		return nil, 0, err
	}

	songPtrs := make([]*dbStructs.Song, len(songs))
	for i := range songs {
		songPtrs[i] = &songs[i]
	}
	return songPtrs, total, nil
}

func (p *Provider) SaveSongs(ctx context.Context, songs []*dbStructs.Song) error {
	return p.repository.SaveSongs(ctx, songs)
}

func (p *Provider) GetSongByID(ctx context.Context, id interface{}) (*dbStructs.Song, error) {
	return p.repository.GetSongByID(ctx, id.(string))
}

func (p *Provider) Close() error {
	return p.client.Close()
}
