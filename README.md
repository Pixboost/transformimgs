# TransformImgs
Image transformations service.

# Requirements

go 1.5+

# Running

$ export GO15VENDOREXPERIMENT=1
$ go get github.com/tools/godep
$ go get github.com/dooman87/transformimgs
$ cd $GOPATH/src/github.com/dooman87/transformimgs
$ godep restore
$ go run cmd/main.go