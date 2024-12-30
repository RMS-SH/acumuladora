# Dockerfile

FROM golang:1.23 AS builder

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copiar os arquivos do projeto (go.mod, go.sum e o restante do código)
COPY go.mod go.sum ./
RUN go mod download

# Copiar o restante dos arquivos
COPY . .

# Compila o binário estático
RUN CGO_ENABLED=0 GOOS=linux go build -o /acumuladora main.go

# ====== Fase final: gerar uma imagem mínima ======
FROM alpine:3.17

# Variáveis de ambiente para Redis (caso sejam necessárias)
ENV REDIS_ADDR=redis:6379
ENV REDIS_PASSWORD=
ENV REDIS_DB=0

# Copiar o binário do build para a imagem final
COPY --from=builder /acumuladora /usr/local/bin/acumuladora

EXPOSE 7511

# Comando padrão de execução
ENTRYPOINT ["/usr/local/bin/acumuladora"]
