
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o bookcabin-api cmd/http/main.go

FROM alpine:latest

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Jakarta

COPY --from=builder /app/bookcabin-api .

COPY --from=builder /app/mock_data ./mock_data

COPY --from=builder /app/docs /app/docs

COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./bookcabin-api"]