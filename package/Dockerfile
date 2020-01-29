FROM alpine:3.11.3
RUN apk -u --no-cache add curl ca-certificates
COPY bin/github-metrics /usr/bin/
CMD ["github-metrics"]
