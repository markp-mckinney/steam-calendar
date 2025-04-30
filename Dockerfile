FROM golang:1.24 AS builder

WORKDIR /build

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -v -o ./steam-calendar

FROM alpine:latest
COPY --from=builder /build/steam-calendar /bin/steam-calendar

CMD ["/bin/steam-calendar"]