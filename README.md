# TransformImgs
Image transformations service.

# Requirements

go 1.5+

# Running

```
$ export GO15VENDOREXPERIMENT=1
$ go get github.com/tools/godep
$ go get github.com/dooman87/transformimgs
$ cd $GOPATH/src/github.com/dooman87/transformimgs
$ godep restore
$ go run cmd/main.go
```

## Docker local - when sources checked out ##

To run server in local environment (using local copy of sources):

```
$ cd docker/local
$ docker build -t transformimgs-local .
$ docker run -it --rm -p 8080:8080 -v //c/Users/dimka//Projects/go/src/github.com/dooman87/transformimgs/://usr/local/go/src/github.com/dooman87/transformimgs/ --name transformimgs-local transformimgs-local
```

## Docker remote - without sources ##

TODO:

# Logging

Service is using [glog](https://github.com/golang/glog) project for logging. It's providing next options that could
be passed to command line:

```
-logtostderr=false
  Logs are written to standard error instead of to files.
-alsologtostderr=false
  Logs are written to standard error as well as to files.
-stderrthreshold=ERROR
  Log events at or above this severity are logged to standard error as well as to files.
-log_dir=""
  Log files will be written to this directory instead of the
  default temporary directory.

Other flags provide aids to debugging.

log_backtrace_at=""
  When set to a file and line number holding a logging statement,
  such as
-log_backtrace_at=gopherflakes.go:234
  a stack trace will be written to the Info log whenever execution
  hits that statement. (Unlike with -vmodule, the ".go" must be
  present.)
-v=0
  Enable V-leveled logging at the specified level.
vmodule=""
  The syntax of the argument is a comma-separated list of pattern=N,
  where pattern is a literal file name (minus the ".go" suffix) or
  "glob" pattern and N is a V level. For instance,
-vmodule=gopher*=3
  sets the V level to 3 in all Go files whose names begin "gopher".
```