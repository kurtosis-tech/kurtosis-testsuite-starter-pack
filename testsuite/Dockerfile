FROM golang:1.15 AS builder
WORKDIR /build
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build Delve
RUN go get github.com/go-delve/delve/cmd/dlv

# Build the application
RUN go build -gcflags="all=-N -l" -o testsuite.bin testsuite/main.go

CMD dlv --listen=:${DEBUGGER_PORT} --headless=true --api-version=2 exec ./testsuite.bin -- \
    `# ------------- Custom variable particular to this test suite -----------------` \
    --api-service-image=${API_SERVICE_IMAGE} \
    --datastore-service-image=${DATASTORE_SERVICE_IMAGE} \
    `# ------------------Kurtosis-specific environment variables -------------------` \
    `# See: https://github.com/kurtosis-tech/kurtosis-docs/blob/master/testsuite-details.md` \
    --kurtosis-api-ip=${KURTOSIS_API_IP} \
    --log-level=${LOG_LEVEL} \
    --metadata-filepath=${METADATA_FILEPATH} \
    --services-relative-dirpath=${SERVICES_RELATIVE_DIRPATH} \
    --test=${TEST}
