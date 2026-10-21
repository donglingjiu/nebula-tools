FROM golang:1.24-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go mod tidy && go build -trimpath -ldflags "-s -w" -o /out/nova .
FROM alpine:3.19
WORKDIR /app
COPY --from=build /out/nova /app/nova
ENV PORT=8080
EXPOSE 8080
VOLUME /app/data
ENTRYPOINT ["/app/nova"]
