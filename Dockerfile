FROM golang:1.24-alpine AS builder

WORKDIR /metachan

RUN apk add --no-cache git make

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:latest

WORKDIR /metachan

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /metachan/bin/metachan .

CMD ["./metachan"]