FROM docker:24.0.7-cli
FROM golang:1.21.4

COPY --from=docker /usr/local/bin/docker /usr/local/bin/
