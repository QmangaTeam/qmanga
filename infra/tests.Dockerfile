FROM golang@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d
RUN apk add --no-cache imagemagick imagemagick-jpeg imagemagick-webp
COPY infra/imagemagick-policy.xml /etc/ImageMagick-7/policy.xml
WORKDIR /src
