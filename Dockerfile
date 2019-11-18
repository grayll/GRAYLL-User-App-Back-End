# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.12 as builder

# Copy local code to the container image.
WORKDIR /go/src/bitbucket.org/grayll/user-app-backend
COPY . .
WORKDIR /home/bc/go/src/github.com/matcornic/hermes
COPY /home/bc/go/src/github.com/matcornic/hermes /home/bc/go/src/github.com/matcornic/hermes
WORKDIR /go/src/bitbucket.org/grayll/user-app-backend
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
COPY --from=builder /go/src/bitbucket.org/grayll/user-app-backend/grayll-app /grayll-app
COPY --from=builder /go/src/bitbucket.org/grayll/user-app-backend/key /key
COPY --from=builder /go/src/bitbucket.org/grayll/user-app-backend/config1.json /config1.json
COPY --from=builder /go/src/bitbucket.org/grayll/user-app-backend/grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json /grayll-app-f3f3f3-firebase-adminsdk-vhclm-e074da6170.json
#ENV PORT 8080
#EXPOSE PORT
# Run the web service on container startup.
CMD ["/grayll-app"]
