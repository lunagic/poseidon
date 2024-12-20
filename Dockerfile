FROM golang:1.23-alpine AS builder
WORKDIR /workspace
COPY . .
RUN go build -ldflags='-s -w' -o build .

FROM debian:latest
COPY --from=builder /workspace/build /usr/local/bin/poseidon
WORKDIR /var/www/html
CMD [ "poseidon" ]
ENV HOST=0.0.0.0
ENV PORT=3000
EXPOSE 3000
