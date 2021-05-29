# pull official base image
FROM golang:1.16

# set working directory
WORKDIR /go/src/app

# add app
COPY . .

# install app dependencies
RUN go get -d -v ./...
RUN go install -v ./...
RUN go build

# start app
CMD ["hivee-core"]