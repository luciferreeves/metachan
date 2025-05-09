FROM golang:1.24 AS builder

WORKDIR /metachan

RUN apt-get update && apt-get install -y gcc libc6-dev make git

ENV CGO_ENABLED=1

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:latest

WORKDIR /metachan

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /metachan/bin/metachan .

CMD ["./metachan"]
