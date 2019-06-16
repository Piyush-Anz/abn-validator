FROM golang as builder

# !!! Docker layer caching will not repeat this step if the repo changes
# !!! You won't be able to build a test copy of your uncommitted code
RUN git clone --branch master https://github.com/bryonbaker/abn-validator.git /go/src/abn-lookup
RUN go get github.com/gorilla/mux
RUN go get github.com/gorilla/handlers
RUN go get github.com/bryonbaker/utils

# vvv Put magic environment variables in this line
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install abn-lookup
# ^^^

# Runtime image
FROM alpine:latest
COPY --from=builder /go/bin/abn-lookup /bin/abn-lookup
ARG VERSION=0.1
RUN echo $VERSION > /image_version
EXPOSE 10000
WORKDIR "/bin"
COPY ./boot-config.json /bin/
COPY ./app-config.json /bin/
CMD ["abn-lookup"]