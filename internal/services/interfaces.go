package services

import (
	"context"

	"github.com/JonayMedina/api-music/internal/structs"
)

// MusicProvider define la interfaz que todos los servicios de música deben implementar
type MusicProvider interface {
	Search(ctx context.Context, query string, artist string, album string) ([]structs.Song, error)
	Name() string
}

// ServiceError representa un error específico del servicio
type ServiceError struct {
	Service string
	Message string
	Err     error
}

func (e *ServiceError) Error() string {
	if e.Err != nil {
		return e.Service + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Service + ": " + e.Message
}

// NewServiceError crea un nuevo error de servicio
func NewServiceError(service, message string, err error) *ServiceError {
	return &ServiceError{
		Service: service,
		Message: message,
		Err:     err,
	}
}
