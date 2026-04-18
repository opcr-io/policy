FROM alpine:3.23.4

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG BUILDPLATFORM

ENV POLICY_FILE_STORE_ROOT=/github/workspace/_policy

RUN echo "BUILDPLATFORM=$BUILDPLATFORM" \
 && echo "TARGETPLATFORM=$TARGETPLATFORM" \
 && echo "TARGETOS=$TARGETOS" \
 && echo "TARGETARCH=$TARGETARCH"

RUN apk add --no-cache bash tzdata ca-certificates

WORKDIR /app

COPY --chmod=755 ${TARGETPLATFORM}/policy /app/
COPY --chmod=755 scripts/build.sh /app/build.sh
COPY --chmod=755 scripts/login.sh /app/login.sh
COPY --chmod=755 scripts/logout.sh /app/logout.sh
COPY --chmod=755 scripts/pull.sh /app/pull.sh
COPY --chmod=755 scripts/push.sh /app/push.sh
COPY --chmod=755 scripts/rm.sh /app/rm.sh
COPY --chmod=755 scripts/save.sh /app/save.sh
COPY --chmod=755 scripts/tag.sh /app/tag.sh

ENTRYPOINT ["./policy"]
