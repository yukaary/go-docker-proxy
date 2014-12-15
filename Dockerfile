FROM golang
#ADD . /go/src/github.com/yukaary/go-docker-proxy
#RUN go get github.com/yukaary/go-docker-proxy
RUN go install github.com/yukaary/go-docker-proxy
