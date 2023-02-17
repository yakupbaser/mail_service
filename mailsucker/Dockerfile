# kullanılacak temel image
FROM golang:alpine AS builder

# gerekli environment değişkenleri
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# uygulama dizini
WORKDIR /app

# go.mod ve go.sum dosyalarını kopyala
COPY go.mod .
COPY go.sum .

# bağımlılıkları yükle
RUN go mod download

# kaynak kodları kopyala
COPY . .

# uygulamayı derle
RUN go build -o mailsucker mailsucker.go

# yeni bir image oluştur
FROM alpine:latest

# uygulama dizini
WORKDIR /app

# derlenmiş uygulamayı kopyala
COPY --from=builder /app/mailsucker .

# varsayılan olarak çalışacak komut
CMD ["./mailsucker"]
