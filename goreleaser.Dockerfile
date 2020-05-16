FROM scratch
COPY go-echoes /
ENTRYPOINT ["/go-echoes"]
