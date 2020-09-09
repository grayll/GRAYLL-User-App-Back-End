# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.12 as builder

# Copy local code to the container image.
WORKDIR /home/bc/go/src/bitbucket.org/huykbc/goautoneg
COPY . /home/bc/go/src/bitbucket.org/huykbc/goautoneg

WORKDIR /go/src/bitbucket.org/grayll/grayll.io-user-app-back-end
COPY . .

# Build the command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
ENV GO111MODULE on
RUN CGO_ENABLED=0 GOOS=linux go build -v -o grayll-app

# Use a Docker multi-stage build to create a lean production image.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM alpine
RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/grayll-app /grayll-app
COPY --from=builder /go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/key /key
COPY --from=builder /go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/config1.json /config1.json
COPY --from=builder /go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/config1-dev.json /config1-dev.json
#COPY --from=builder /go/src/bitbucket.org/grayll/grayll.io-user-app-back-end/grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json /grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip
#ENV PORT 8080
#EXPOSE PORT
# Run the web service on container startup.
CMD ["/grayll-app"]
