# build stage
FROM golang:alpine AS build-env
ADD . /app
RUN cd /app && go get ./... && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o echoes

# final stage
FROM centurylink/ca-certs
COPY --from=build-env /app/echoes /
EXPOSE 54321
ENTRYPOINT ["/echoes"]