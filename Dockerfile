
FROM golang

WORKDIR /go/src/github.com/patrickjmcd/lake-info
ADD main.go .
RUN go get -v -d
RUN go install github.com/patrickjmcd/lake-info

ENTRYPOINT /go/bin/lake-info
