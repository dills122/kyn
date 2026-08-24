FROM gcr.io/distroless/static:nonroot

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/kyn /usr/local/bin/kyn

ENTRYPOINT ["/usr/local/bin/kyn"]
