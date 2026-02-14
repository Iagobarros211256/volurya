# ---------- BUILD STAGE ----------
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Copiar dependências primeiro (cache)
COPY go.mod go.sum ./
RUN go mod download

# Copiar código
COPY . .

# Compilar binário estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd


# ---------- FINAL STAGE ----------
FROM alpine:latest

WORKDIR /app

# Copiar apenas o binário
COPY --from=builder /app/app .
COPY --from=builder /app/views ./views


EXPOSE 8080

CMD ["./app"]
