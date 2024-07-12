FROM gcr.io/distroless/static:nonroot
WORKDIR /
ENV TERM=xterm-256color
COPY cli-of-life /
ENTRYPOINT ["/cli-of-life"]
