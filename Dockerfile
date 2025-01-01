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

# Ensure that the Chromium binary is in the PATH
ENV PATH="/usr/lib/chromium/:${PATH}"

# Command to run the executable
CMD ["./main"]
