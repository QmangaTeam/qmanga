FROM cgr.dev/chainguard/minio@sha256:c93745c56fccdaf0e5dcb88999a96a65a136428beb045ef5d546255a80d747a7 AS upstream
FROM alpine@sha256:ce64758a109eb420d874a118f87920e625e12d3634e03b4a5573fd9f6e5d3507
RUN apk add --no-cache ca-certificates && addgroup -S minio && adduser -S -G minio minio && mkdir /data && chown minio:minio /data
COPY --from=upstream /usr/bin/minio /usr/local/bin/minio
USER minio
ENTRYPOINT ["minio"]
CMD ["server", "/data", "--console-address", ":9001"]

