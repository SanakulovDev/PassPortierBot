# 1-bosqich: Qurish (Build stage)
FROM golang:1.24-alpine AS builder

# Ishchi katalogni yaratish
WORKDIR /app

# Zaruriy paketlarni o'rnatish
RUN apk add --no-cache git

# Go modullarni yuklab olish
COPY go.mod go.sum ./
RUN go mod download

# Loyiha kodlarini nusxalash
COPY . .

RUN go mod tidy

# Binar faylni kompilyatsiya qilish (Static link orqali)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o passportier ./cmd/main.go

# 2-bosqich: Ishga tushirish (Final stage)
FROM alpine:latest

# SSL sertifikatlarni o'rnatish (API so'rovlar uchun shart)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Builder bosqichidan faqat binar faylni nusxalash
COPY --from=builder /app/passportier .
# .env faylini nusxalash (agar mavjud bo'lsa)
COPY .env .

# Ijro huquqini berish
RUN chmod +x ./passportier

# Botni ishga tushirish
CMD ["./passportier"]