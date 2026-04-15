FROM golang:1.26-alpine AS builder
# Build only backend module
WORKDIR /src/backend
COPY backend/go.mod ./go.mod
COPY backend/ ./
# ensure module dependencies for code added (e.g. oauth2, jwt)
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /qmanga .

FROM alpine:3.18
COPY --from=builder /qmanga /qmanga
# copy migrations directly from build context
COPY migrations /migrations
WORKDIR /
EXPOSE 8080
ENTRYPOINT ["/qmanga"]
