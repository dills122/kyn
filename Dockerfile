FROM gcr.io/distroless/static:nonroot

COPY kyn /usr/local/bin/kyn

ENTRYPOINT ["/usr/local/bin/kyn"]
