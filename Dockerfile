# Build Stage
ARG GO_VERSION
FROM golang:$GO_VERSION-alpine AS build
RUN apk add --no-cache bash build-base git tree curl protobuf openssh
WORKDIR /src

# make sure git ssh is properly setup so we can access private repos
RUN mkdir -p $HOME/.ssh && umask 0700 \
	&& git config --global url."git@github.com:".insteadOf https://github.com/ \
	&& ssh-keyscan github.com >> $HOME/.ssh/known_hosts

ENV GOBIN=/bin
ENV ROOT_DIR=/src

# generate & build
ARG VERSION
ARG COMMIT
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
		--mount=type=cache,target=/root/.cache/go-build \
		--mount=type=ssh \
		go run mage.go deps build

FROM alpine:3
ARG VERSION
ARG COMMIT
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.source=https://github.com/aserto-dev/policy
LABEL org.opencontainers.image.title="policy"
LABEL org.opencontainers.image.revision=$COMMIT
LABEL org.opencontainers.image.url="https://aserto.com"

RUN apk add --no-cache bash
WORKDIR /app
COPY --from=build /src/dist/build_linux_amd64/policy /app/

ENTRYPOINT ["./policy"]
