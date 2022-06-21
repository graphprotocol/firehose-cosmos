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

ARG USER_ID=1234

WORKDIR /app/

COPY --from=build /build/build/firehose-cosmos /app/firehose

RUN addgroup --gid ${USER_ID} firehose && \
    adduser --system --uid ${USER_ID} firehose && \
    chown -R firehose:firehose /app

USER ${USER_ID}
