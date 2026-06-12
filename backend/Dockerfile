FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o server ./cmd/server

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server /app/server
# Placeholder until Firebase credentials are added in Phase 3
RUN echo '{}' > /app/firebase-key.json
EXPOSE 8080
CMD ["/app/server"]
