[![Build Status](https://travis-ci.org/danielfireman/tsmgo.svg?branch=master)](https://travis-ci.org/danielfireman/tsmgo) [![Coverage Status](https://codecov.io/gh/danielfireman/tsmgo/branch/master/graph/badge.svg)](https://codecov.io/gh/danielfireman/tsmgo/branch/master/graph/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/danielfireman/tsmgo)](https://goreportcard.com/report/github.com/danielfireman/tsmgo) [![GoDoc](https://godoc.org/github.com/danielfireman/tsmgo?status.svg)](https://godoc.org/github.com/danielfireman/tsmgo)

# tsmgo

#golang library which makes intuitive to work with time-series data in Mongo DB.

The schema design is inspired by [this article](https://www.mongodb.com/blog/post/schema-design-for-time-series-data-in-mongodb).

# Using tsmgo

```go
import "github.com/danielfireman/tsmgo"
```

**New tsmgo.Session**

```go
tsSession := tsmgo.NewSession(mgoSession)
c, _ := tsSession.C(dbName, colName)
```

or

```go
tsSession, err := tsmgo.Dial(mongoURL)
c, _ := tsSession.C(dbName, colName)
```

Where mgoSession is a [github.com/globalsign/mgo#Session](https://godoc.org/github.com/globalsign/mgo#Session).

**Adding timeseries records**

```go
c.Upsert(myField, TSRecord{time.Now(), 1})
```

or

```go
t1 := time.Now()
t2 := t1.Add(10 * time.Second)
c.Upsert(myField, TSRecord{t1, 1}, TSRecord{t2, 2})
```

**Retrieving last item inserted**
```go
last, _ := tsmgoC.Last(myField)
```

**Retrieving all itens inserted in the last 24 hours**
```go
now := time.Now()
items, _ := tsmgoC.Interval(myField, now.Add(-24*time.Hour), now)
```

# Contributing

1. Install [dep](https://github.com/golang/dep/blob/master/docs/installation.md)
1. `dep ensure`
1. `go test -v`

Either if you're fixing a bug or adding a new feature, please add a test to cover it.

If all tests passes, please you're ready to send the PR.
