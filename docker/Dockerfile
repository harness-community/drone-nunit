FROM debian:bookworm
ENV GODEBUG=netdns=go

# Install required dependencies
RUN apt-get update && apt-get install -y \
    libxml2 \
    libxslt1.1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the plugin and the stylesheet from the release folder to the Docker image
COPY release/linux/amd64/plugin /bin/plugin
COPY /docs/conversionStyleSheet.xsl /docs/conversionStyleSheet.xsl

# Set the entrypoint for the Docker container
ENTRYPOINT ["/bin/plugin"]