FROM node:14.17 AS builder

WORKDIR /build

COPY package.json .
COPY yarn.lock .

RUN yarn install

# Copy local static files inside the testsuite image
COPY testsuite/static_files /static-files

# Copy the code into the container
COPY . .
RUN yarn build 

# Build the application
CMD node build/main.js
