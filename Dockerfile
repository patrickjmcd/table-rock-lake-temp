
FROM golang

WORKDIR /go/src/github.com/patrickjmcd/table-rock-lake-temp
ADD main.go .
RUN go get -v -d
RUN go install github.com/patrickjmcd/table-rock-lake-temp

ENTRYPOINT /go/bin/table-rock-lake-temp
