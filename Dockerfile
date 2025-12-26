FROM golang:1.25.4

WORKDIR /go/src/app

# Copiar go.mod/go.sum primeiro para cache
COPY go.mod go.sum ./
RUN go mod download

# Copiar o resto do código
COPY . .

# Instalar Air
RUN  go install github.com/air-verse/air@latest

EXPOSE 8000

# Rodar Air com hot reload
CMD ["air", "-c", ".air.toml"]
