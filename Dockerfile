FROM golang:1.24-alpine AS build
ENV GOTOOLCHAIN=auto
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/nova .
FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/nova /app/nova
ENV PORT=8080 NOVA_DATA_DIR=/app/data
EXPOSE 8080
VOLUME /app/data
ENTRYPOINT ["/app/nova"]
