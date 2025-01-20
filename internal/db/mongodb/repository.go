// internal/db/mongodb/repository.go
package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/JonayMedina/api-music/internal/structs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const songCollection = "songs"

type SongRepository struct {
	client *MongoClient
	coll   *mongo.Collection
}

func NewSongRepository(client *MongoClient) *SongRepository {
	return &SongRepository{
		client: client,
		coll:   client.Collection(songCollection),
	}
}

// CreateIndexes crea los índices necesarios para la colección de canciones
func (r *SongRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "artist", Value: "text"},
				{Key: "album", Value: "text"},
			},
			Options: options.Index().SetName("text_search"),
		},
		{
			Keys:    bson.D{{Key: "origin", Value: 1}},
			Options: options.Index().SetName("origin_idx"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetName("created_at_idx"),
		},
	}

	_, err := r.coll.Indexes().CreateMany(ctx, indexes)
	return err
}

// SaveSongs guarda un conjunto de canciones en la base de datos
func (r *SongRepository) SaveSongs(ctx context.Context, songs []structs.Song) error {
	if len(songs) == 0 {
		return nil
	}

	now := time.Now().Unix()
	documents := make([]interface{}, len(songs))

	for i, song := range songs {
		song.CreatedAt = now
		song.UpdatedAt = now
		documents[i] = song
	}

	opts := options.InsertMany().SetOrdered(false)
	_, err := r.coll.InsertMany(ctx, documents, opts)
	return err
}

// SearchSongs busca canciones según los criterios especificados
func (r *SongRepository) SearchSongs(ctx context.Context, query string, artist string, album string, page int, limit int) ([]structs.Song, int64, error) {
	filter := bson.M{}

	// Construir filtro de búsqueda
	if query != "" || artist != "" || album != "" {
		searchTerms := []string{}
		if query != "" {
			searchTerms = append(searchTerms, query)
		}
		if artist != "" {
			searchTerms = append(searchTerms, artist)
		}
		if album != "" {
			searchTerms = append(searchTerms, album)
		}

		filter["$text"] = bson.M{
			"$search": strings.Join(searchTerms, " "), // Join search terms with space
		}
	}

	// Opciones de paginación
	skip := (page - 1) * limit
	findOptions := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))

	// Ejecutar búsqueda
	cursor, err := r.coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Obtener total de documentos
	total, err := r.coll.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Decodificar resultados
	var songs []structs.Song
	if err = cursor.All(ctx, &songs); err != nil {
		return nil, 0, err
	}

	return songs, total, nil
}

// GetSongByID obtiene una canción por su ID
func (r *SongRepository) GetSongByID(ctx context.Context, id string) (*structs.Song, error) {
	var song structs.Song
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&song)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &song, nil
}

// DeleteOldSongs elimina canciones más antiguas que la fecha especificada
func (r *SongRepository) DeleteOldSongs(ctx context.Context, before time.Time) error {
	filter := bson.M{
		"created_at": bson.M{
			"$lt": before.Unix(),
		},
	}

	_, err := r.coll.DeleteMany(ctx, filter)
	return err
}
