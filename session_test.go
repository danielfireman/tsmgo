package tsmgo

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/globalsign/mgo/dbtest"
	"github.com/matryer/is"
)

var mgoSession *mgo.Session
var mongoDB dbtest.DBServer

func TestMain(m *testing.M) {
	tempDir, _ := ioutil.TempDir("", "predictor_testing")
	defer func() { os.RemoveAll(tempDir) }()
	mongoDB.SetPath(tempDir)
	retCode := m.Run()
	mongoDB.Stop()
	os.Exit(retCode)
}

const (
	dbName  = "db"
	colName = "col"
	type1   = "type"
)

func ExampleCollection() {
	mgoSession = mongoDB.Session()
	defer mongoDB.Wipe()
	defer mgoSession.Close()

	s := NewSession(mgoSession)
	defer s.Close()

	myField := "myfield"
	t1 := time.Now()
	t2 := t1.Add(10 * time.Second)

	tsmgoC, _ := s.C(dbName, colName)
	tsmgoC.Upsert(myField, TSRecord{t1, 1}, TSRecord{t2, 2})

	last, _ := tsmgoC.Last(myField)
	fmt.Println(last.Value)

	records, _ := tsmgoC.Interval(myField, t1, t2)
	sort.Sort(InverseChronologicalOrdering(records))
	fmt.Println(records[0].Value, records[1].Value)

	// Output: 2
	// 2 1
}

func TestCollection_TSUpsert_oneRecord(t *testing.T) {
	mgoSession = mongoDB.Session()
	defer mongoDB.Wipe()
	defer mgoSession.Close()

	is := is.New(t)
	s := NewSession(mgoSession)
	defer s.Close()

	tsmgoC, err := s.C(dbName, colName)
	is.NoErr(err) // s.C(dbName, colName)
	t1 := time.Now()
	res1, err := tsmgoC.Upsert(type1, TSRecord{t1, 1})
	is.NoErr(err)              // tsmgoC.TSUpsert(type1, TSRecord{t1, 1}) error
	is.Equal(1, res1.Matched)  // res1.Matched
	is.Equal(0, res1.Modified) // res1.Modified

	var records []TSRecord
	mgoC := mgoSession.DB(dbName).C(colName)
	is.NoErr(mgoC.Find(bson.M{typeField: type1}).All(&records)) // mgoC.Find.All error
	is.Equal(1, len(records))                                   // len(records)
	is.Equal(records[0].Timestamp.Unix(), t1.In(time.UTC).Unix())
	is.Equal(records[0].Value.(int), 1)
}

func TestCollection_TSUpsert_override(t *testing.T) {
	mgoSession = mongoDB.Session()
	defer mongoDB.Wipe()
	defer mgoSession.Close()

	is := is.New(t)
	s := NewSession(mgoSession)
	defer s.Close()

	tsmgoC, err := s.C(dbName, colName)
	is.NoErr(err) // s.C(dbName, colName)
	t1 := time.Now()
	res1, err := tsmgoC.Upsert(type1, TSRecord{t1, 1})
	is.NoErr(err)              // tsmgoC.TSUpsert(type1, TSRecord{t1, 1}) error
	is.Equal(1, res1.Matched)  // res1.Matched
	is.Equal(0, res1.Modified) // res1.Modified

	res2, err := tsmgoC.Upsert(type1, TSRecord{t1, 2})
	is.NoErr(err)              // tsmgoC.TSUpsert(type1, TSRecord{t1, 2}) error
	is.Equal(1, res2.Matched)  // res2.Matched
	is.Equal(1, res2.Modified) // res2.Modified

	var records []TSRecord
	mgoC := mgoSession.DB(dbName).C(colName)
	is.NoErr(mgoC.Find(bson.M{typeField: type1}).All(&records)) // mgoC.Find.All error
	is.Equal(1, len(records))                                   // len(records)
	is.Equal(records[0].Timestamp.Unix(), t1.In(time.UTC).Unix())
	is.Equal(records[0].Value.(int), 2)
}

func TestCollection_TSUpsert_multipleRecords(t *testing.T) {
	mgoSession = mongoDB.Session()
	defer mongoDB.Wipe()
	defer mgoSession.Close()

	is := is.New(t)
	s := NewSession(mgoSession)
	defer s.Close()

	tsmgoC, err := s.C(dbName, colName)
	is.NoErr(err) // s.C(dbName, colName)
	t1 := time.Now()
	t2 := t1.Add(10 * time.Second)
	res1, err := tsmgoC.Upsert(type1, TSRecord{t1, 1}, TSRecord{t2, 2})
	is.NoErr(err)              // tsmgoC.TSUpsert(type1, TSRecord{t1, 1}, TSRecord{t2, 2}) error
	is.Equal(2, res1.Matched)  // res1.Matched
	is.Equal(0, res1.Modified) // res1.Modified

	type2 := "type2"
	res2, err := tsmgoC.Upsert(type2, TSRecord{t1, 3})
	is.NoErr(err)              // tsmgoC.TSUpsert(type1, TSRecord{t1, 1}, TSRecord{t2, 2}) error
	is.Equal(1, res2.Matched)  // res2.Matched
	is.Equal(0, res2.Modified) // res2.Modified

	mgoC := mgoSession.DB(dbName).C(colName)

	var rec1 []TSRecord
	is.NoErr(mgoC.Find(bson.M{typeField: type1}).All(&rec1)) // mgoC.Find("type":"type1").All() error
	is.Equal(2, len(rec1))                                   // len(rec1)
	is.Equal(rec1[0].Timestamp.Unix(), t1.In(time.UTC).Unix())
	is.Equal(rec1[0].Value.(int), 1)
	is.Equal(rec1[1].Timestamp.Unix(), t2.In(time.UTC).Unix())
	is.Equal(rec1[1].Value.(int), 2)

	var rec2 []TSRecord
	is.NoErr(mgoC.Find(bson.M{typeField: type2}).All(&rec2)) // mgoC.Find("type":"type2").All error
	is.Equal(1, len(rec2))                                   // len(rec2)
	is.Equal(rec2[0].Timestamp.Unix(), t1.In(time.UTC).Unix())
	is.Equal(rec2[0].Value.(int), 3)
}

func TestCollection_TSLast(t *testing.T) {
	mgoSession = mongoDB.Session()
	defer mongoDB.Wipe()
	defer mgoSession.Close()

	is := is.New(t)
	s := NewSession(mgoSession)
	defer s.Close()

	tsmgoC, err := s.C(dbName, colName)
	is.NoErr(err) // s.C(dbName, colName)
	t1 := time.Now()
	t2 := t1.Add(10 * time.Second)
	res1, err := tsmgoC.Upsert(type1, TSRecord{t1, 1}, TSRecord{t2, 2})
	is.NoErr(err)              // tsmgoC.TSUpsert(type1, TSRecord{t1, 1}, TSRecord{t2, 2}) error
	is.Equal(2, res1.Matched)  // res1.Matched
	is.Equal(0, res1.Modified) // res1.Modified

	rec, err := tsmgoC.Last(type1)
	is.NoErr(err) // tsmgoC.Last(type1) error
	is.Equal(rec.Timestamp.Unix(), t2.In(time.UTC).Unix())
	is.Equal(rec.Value.(int), 2)
}

func TestCollection_Interval(t *testing.T) {
	mgoSession = mongoDB.Session()
	defer mongoDB.Wipe()
	defer mgoSession.Close()

	is := is.New(t)
	s := NewSession(mgoSession)
	defer s.Close()

	tsmgoC, err := s.C(dbName, colName)
	is.NoErr(err) // s.C(dbName, colName)
	t1 := time.Now()
	t2 := t1.Add(10 * time.Second)
	res1, err := tsmgoC.Upsert(type1, TSRecord{t1, 1}, TSRecord{t2, 2})
	is.NoErr(err)              // tsmgoC.TSUpsert(type1, TSRecord{t1, 1}, TSRecord{t2, 2}) error
	is.Equal(2, res1.Matched)  // res1.Matched
	is.Equal(0, res1.Modified) // res1.Modified

	rec, err := tsmgoC.Interval(type1, t1, t2)
	is.NoErr(err) // tsmgoC.Interval(t1, t2)
	is.Equal(rec[0].Timestamp.Unix(), t1.In(time.UTC).Unix())
	is.Equal(rec[0].Value.(int), 1)
	is.Equal(rec[1].Timestamp.Unix(), t2.In(time.UTC).Unix())
	is.Equal(rec[1].Value.(int), 2)
}
