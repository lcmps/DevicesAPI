# Build stage
FROM golang:1.25.1 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -trimpath -o devices -v .

# Final stage
FROM scratch

WORKDIR /app
COPY --from=builder /app/devices .

ENV PORT=9001
EXPOSE 9001

ENTRYPOINT ["/app/devices"]