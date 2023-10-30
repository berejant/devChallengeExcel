FROM golang:1.21-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
# Build the binary.
RUN go build -ldflags="-w -s" -tags=nomsgpack -o /app .

RUN cat /etc/passwd | grep nobody > /etc/passwd.nobody

# build a small image
FROM alpine

COPY --from=builder /etc/passwd.nobody /etc/passwd
COPY --from=builder /app /app

ENV DATABASE_FILEPATH /storage/sheets.db
VOLUME /storage
RUN mkdir /storage && touch /storage/sheets.db && chmod 777 -R /storage

# Run
USER nobody

ENTRYPOINT ["/app"]
