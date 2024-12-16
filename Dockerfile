FROM golang:1.23-alpine3.20 AS build

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /app/bin/anu ./cmd/anu

FROM alpine:3.20

RUN apk -U add ca-certificates mailcap

COPY --from=build /app/bin/anu /app/bin/anu

EXPOSE 9200

CMD ["/app/bin/anu"]
