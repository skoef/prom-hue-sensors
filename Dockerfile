FROM alpine:latest

COPY tempread /bin/tempread

CMD ["/bin/tempread"]
