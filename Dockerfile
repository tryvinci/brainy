FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/brainy-api ./cmd/api
RUN CGO_ENABLED=0 go build -o /bin/brainy-worker ./cmd/worker

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /bin/brainy-api /bin/brainy-worker /app/
COPY packs /app/packs
ENV BRAINY_HTTP_ADDR=:8080
EXPOSE 8080
CMD ["/app/brainy-api"]
