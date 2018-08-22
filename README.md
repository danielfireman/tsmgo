[![Build Status](https://travis-ci.org/danielfireman/tsmgo.svg?branch=master)](https://travis-ci.org/danielfireman/tsmgo) [![Coverage Status](https://codecov.io/gh/danielfireman/tsmgo/branch/master/graph/badge.svg)](https://codecov.io/gh/danielfireman/tsmgo/branch/master/graph/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/danielfireman/tsmgo)](https://goreportcard.com/report/github.com/danielfireman/tsmgo) [![GoDoc](https://godoc.org/github.com/danielfireman/tsmgo?status.svg)](https://godoc.org/github.com/danielfireman/tsmgo)

# tsmgo

> Disclaimer: a lot in flux

Golang library which makes easier to work with timeseries data in Mongo DB

# Using tsmgo

**Adding timeseries records**

```go
tsSession := NewSession(mgoSession)
tsC, _ := tsSession.C(dbName, colName)
tsC.TSUpsert(myField, TSRecord{time.Now(), 1})
```

**Retrieving last item inserted**
```go
tsSession := NewSession(mgoSession)
last, _ := tsmgoC.Last(myField)
```

**Retrieving all itens inserted in the last 24 hours**
```go
tsSession := NewSession(mgoSession)
now := time.Now()
items, _ := tsmgoC.Interval(myField, now.Add(-24*time.Hour), now)
```

Where mgoSession is a [github.com/globalsign/mgo#Session](https://godoc.org/github.com/globalsign/mgo#Session).

# Contributing

1. Install [dep](https://github.com/golang/dep/blob/master/docs/installation.md)
1. `dep ensure`
1. `go test -v`

Either if you're fixing a bug or adding a new feature, please add a test to cover it.

If all tests passes, please you're ready to send the PR.
