# Etapa de build
FROM golang:1.23 AS builder

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos do projeto para dentro do container
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o fullcycle-rate-limiter main.go

# Etapa final: imagem para execução
FROM alpine:latest

WORKDIR /root/

# Instalar dependências para execução
RUN apk --no-cache add ca-certificates

# Copiar o binário da etapa de build para a imagem final
COPY --from=builder /app/fullcycle-rate-limiter .

# Copiar o arquivo de configuração .env
COPY .env .env

# Expõe a porta para o servidor
EXPOSE 8080

# Comando de execução da aplicação
CMD ["./fullcycle-rate-limiter"]
