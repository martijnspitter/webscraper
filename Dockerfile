# Use a lightweight base image
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
    chromium-chromedriver

# Set the working directory in the container
WORKDIR /app

# Copy the pre-built binary into the container
COPY main .

# Make sure the binary is executable
RUN chmod +x main

# Ensure that the Chromium binary is in the PATH
ENV PATH="/usr/lib/chromium/:${PATH}"

# Verify Chromium installation
RUN chromium-browser --version

# Command to run the executable
CMD ["./main"]
