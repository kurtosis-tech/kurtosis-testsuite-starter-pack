FROM node:14.17.1-alpine as builder 

# We disable CGO here due to:
# 1) https://github.com/golang/go/issues/28065 that prevents 'go test' from running inside an Alpine container
# 2) https://stackoverflow.com/questions/36279253/go-compiled-binary-wont-run-in-an-alpine-docker-container-on-ubuntu-host which
#       which prevents from just switching to the Buster build image
# Sadly, this is slower: https://stackoverflow.com/questions/47714278/why-is-compiling-with-cgo-enabled-0-slower
ENV CGO_ENABLED=0

WORKDIR /build
# Copy and download dependencies using package.json
COPY package.json .
RUN npm install 
#TODO - is npm install enough or do I need to ensure it is installed through below commands
#RUN brew install nvm
#RUN nvm install 14.17.1
#RUN npm install -g typescript

# Copy the code into the container
COPY . .

RUN npm run build
#We would need to add a build section to our package.json for the build to run

# Build the application
#RUN GOOS=linux go build -o testsuite.bin testsuite/main.go

# # ============= Execution Stage ================
# FROM alpine:3.12 AS execution

# WORKDIR /run

# # Copy the code into the container
# COPY --from=builder /build/testsuite.bin .

# # Copy local static files inside the testsuite image
# COPY testsuite/static_files /static-files

# CMD ./testsuite.bin
