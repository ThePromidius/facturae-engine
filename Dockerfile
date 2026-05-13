# ---- Build Stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copiar el manifiesto de dependencias
COPY go.mod go.sum* ./
RUN go mod download

# Copiar todo el código fuente
COPY . .

# Compilar un binario estático optimizado (sin información de debug para que pese menos)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o facturae-engine ./src/cmd/facturae-engine

# ---- Run Stage ----
FROM alpine:latest

WORKDIR /app

# Instalar dependencias clave del sistema operativo para nuestro motor:
# - openssl: Requerido por nuestro `internal/signing` para procesar archivos .p12
# - libxml2-utils: Requerido por `internal/schema/validator.go` (xmllint) para la validación XSD
# - tzdata: Buenas prácticas para el manejo de fechas y zonas horarias en Go
RUN apk add --no-cache openssl libxml2-utils tzdata

# Copiar el binario desde la fase de construcción
COPY --from=builder /app/facturae-engine .

# Crear carpetas necesarias para la persistencia y caché de esquemas
RUN mkdir -p /app/data /app/schemas

# Exponer el puerto por el que escucha nuestro Sidecar HTTP
EXPOSE 8080

# Definir el comando de arranque por defecto
# Por defecto lo arrancamos sin envío a la AEAT (para desarrollo). En prod se sobreescribirá.
CMD ["./facturae-engine", "-aeat=false", "-schemas=/app/schemas"]
