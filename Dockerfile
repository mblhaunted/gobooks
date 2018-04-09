FROM  golang:1.10.1
RUN   curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
COPY  . /go/src/github.com/mblhaunted/gobooks
WORKDIR /go/src/github.com/mblhaunted/gobooks
