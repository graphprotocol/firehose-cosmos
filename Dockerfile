# ------------------------------------------------------------------------------
# Go Builder Image
# ------------------------------------------------------------------------------
FROM golang:1.18 AS build

ENV CGO_ENABLED=0

WORKDIR /build
COPY . .

RUN go mod download
RUN make build

# ------------------------------------------------------------------------------
# Target Image
# ------------------------------------------------------------------------------
FROM alpine AS release

WORKDIR /app/

COPY --from=build /build/build/firehose-cosmos /app/firehose

RUN addgroup --gid 1234 firehose
RUN adduser --system --uid 1234 firehose
RUN chown -R firehose:firehose /app

USER 1234
