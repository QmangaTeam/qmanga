FROM golang@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/admin ./cmd/admin
FROM alpine@sha256:ce64758a109eb420d874a118f87920e625e12d3634e03b4a5573fd9f6e5d3507
RUN apk add --no-cache ca-certificates && addgroup -S qmanga && adduser -S -G qmanga qmanga
COPY --from=build /out/api /usr/local/bin/api
COPY --from=build /out/admin /usr/local/bin/qmanga-admin
USER qmanga
EXPOSE 8080
ENTRYPOINT ["api"]

