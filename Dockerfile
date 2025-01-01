FROM alpine:3.14

# Install necessary packages
RUN apk --no-cache add \
    ca-certificates \
    curl \
    bash \
    chromium \
    nss \
    freetype \
    harfbuzz \
    ttf-freefont \
    udev \
    ttf-dejavu \
    libstdc++ \
    chromium-chromedriver \
    sqlite

WORKDIR /app

# Copy binary and .env from builder
COPY main .
COPY .env .

# Make sure the binary is executable
RUN chmod +x main

# Set Chrome environment variables
ENV CHROME_BIN=/usr/bin/chromium-browser
ENV CHROME_PATH=/usr/lib/chromium/
# Add chromium to PATH
ENV PATH="/usr/lib/chromium/:/usr/bin/chromium-browser:${PATH}"

# Verify Chrome installation
RUN chromium-browser --version || true

# Command to run the executable
CMD ["./main"]
