FROM golang:1.21-alpine

WORKDIR /app

# Instalar dependencias del sistema
RUN apk add --no-cache git curl

# Copiar archivos de dependencias
COPY go.mod go.sum ./

# Descargar dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -o /music-api ./cmd/api

# Exponer puerto
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["/music-api"]