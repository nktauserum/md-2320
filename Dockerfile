# do not set latest tag in production!
FROM golang:latest as builder 

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o md2320 ./cmd/main.go

FROM python:latest 

WORKDIR /app
COPY --from=builder /build/md2320 /app/md2320
RUN python3 -m pip install -U yt-dlp
RUN chmod +x /app/md2320
CMD [ "/app/md2320" ]