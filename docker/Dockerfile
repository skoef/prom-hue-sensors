FROM --platform=$TARGETPLATFORM alpine:latest
ARG TARGETPLATFORM

LABEL maintainer="Reinier Schoof <reinier@skoef.nl>"

WORKDIR /
USER nobody

COPY out/${TARGETPLATFORM}/prom-hue-sensors bin/prom-hue-sensors

CMD ["/bin/prom-hue-sensors"]
