FROM golang:1.24-alpine AS backend_builder
RUN apk add --no-cache git gcc g++
WORKDIR /workspace
COPY . .
RUN git clean -Xdff
RUN CGO_ENABLED=1 go build -ldflags='-s -w' -o /usr/local/bin/build .

FROM alpine:latest
WORKDIR /workspace
COPY --from=backend_builder /usr/local/bin/build /usr/local/bin/build
CMD [ "build" ]
ENV HOST=0.0.0.0
ENV PORT=3000
EXPOSE 3000
