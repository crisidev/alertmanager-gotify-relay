#
# build container
#
FROM quay.io/armswarm/alpine:3.7
RUN apk --no-cache add ca-certificates curl

COPY config.yml /config.yml
COPY alertmanager-gotify-relay /alertmanager-gotify-relay

EXPOSE     8001
ENTRYPOINT [ "/alertmanager-gotify-relay" ]
