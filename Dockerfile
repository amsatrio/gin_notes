FROM golang:latest as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/gin_notes_release .



FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/bin/gin_notes_release .
COPY --from=builder /app/.env_docker .
COPY --from=builder /app/cert cert

RUN mv .env_docker .env

# EXPOSE 8802

CMD ["./gin_notes_release"] 