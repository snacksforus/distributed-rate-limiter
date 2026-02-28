FROM alpine:latest AS test

RUN apk add --no-cache go

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd ./cmd/
COPY api ./api/
COPY internal ./internal/

CMD ["go", "test", "-v", "-race", "-cover", "./..."]

FROM alpine:latest AS builder

RUN apk add --no-cache go

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY cmd ./cmd/
COPY api ./api/
COPY internal ./internal/

RUN go build -o api-server cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-server .

EXPOSE 8080

CMD ["./api-server"]