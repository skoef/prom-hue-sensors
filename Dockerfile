FROM --platform=$TARGETPLATFORM alpine:latest

ARG TARGETPLATFORM

COPY out/${TARGETPLATFORM}/tempread /bin/tempread

CMD ["/bin/tempread"]
