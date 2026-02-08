FROM alpine:latest AS builder

RUN apk add --no-cache go

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./
COPY internal ./internal/
COPY api ./api/

RUN go build -o api-server main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api-server .

EXPOSE 8080

CMD ["./api-server"]