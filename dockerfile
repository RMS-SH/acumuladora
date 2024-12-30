# Escolhe uma imagem base oficial do Golang com versão estável (ex: 1.20)
FROM golang:1.23 AS builder

# Definir o diretório de trabalho dentro do container
WORKDIR /app

# Copiar os arquivos do projeto (go.mod, go.sum e o restante do código)
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante dos arquivos (main.go, pastas config, domain, infrastructure, interfaces, repository e usecase)
COPY . .

# Compila o binário estático para o serviço requisicao
RUN CGO_ENABLED=0 GOOS=linux go build -o /filaflowise main.go

# ====== Fase final: gerar uma imagem mínima ======
FROM alpine:3.17

# Definir variável de ambiente para configuração (se necessário)
ENV REDIS_URL=redis://redis:6379
ENV MONGO_URI=mongodb://rms:2ce8373f618edcda7557ea61f3566d88@mongodb:27017/?authSource=admin&readPreference=primary&ssl=false&directConnection=true

# Copiar o binário do build para a imagem final
COPY --from=builder /filaflowise /usr/local/bin/filaflowise

# Expor a porta 9990 (onde a aplicação escuta)
EXPOSE 9988

# Comando padrão de execução
ENTRYPOINT ["/usr/local/bin/filaflowise"]
