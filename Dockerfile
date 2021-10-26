# Build Stage
ARG GO_VERSION
FROM golang:$GO_VERSION-alpine AS build
RUN apk add --no-cache bash build-base git tree curl protobuf openssh
WORKDIR /src

ENV GOBIN=/bin
ENV ROOT_DIR=/src

# generate & build
ARG VERSION
ARG COMMIT
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
		--mount=type=cache,target=/root/.cache/go-build \
		go run mage.go deps build

FROM alpine:3
ARG VERSION
ARG COMMIT
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.source=https://github.com/opcr-io/policy
LABEL org.opencontainers.image.title="policy"
LABEL org.opencontainers.image.revision=$COMMIT
LABEL org.opencontainers.image.url="https://openpolicyregistry.io"

RUN apk add --no-cache bash
WORKDIR /app
COPY --from=build /src/dist/build_linux_amd64/policy /app/

COPY --from=build /src/scripts /app/
RUN  chmod +x /app/*.sh

ENV POLICY_FILE_STORE_ROOT=/github/workspace/_policy

ENTRYPOINT ["./policy"]
