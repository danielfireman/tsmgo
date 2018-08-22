package tsmgo

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	timestampIndexField = "timestamp"  // Should match TSRecord.Timestamp bson field name.
	typeField           = "type"
	valueField          = "value"
)

// Session represents a connection to a mongo timeseries collection.
type Session struct {
	*mgo.Session
}

// TSCopy works just like New, but preserves the database and any authentication
// information from the original session.
func (s *Session) TSCopy() *Session {
	return NewSession(s.Copy())
}

// NewSession creates a new Session instance based on a copy of the passed-in instance. This returned allows
// the communication to the underlying timeseries mongo database.
func NewSession(s *mgo.Session) *Session {
	return &Session{s.Copy()}
}

// Dial sets up a connection to the specified timeseries database specified by the passed-in URI.
func Dial(uri string) (*Session, error) {
	info, err := mgo.ParseURL(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid db URI:\"%s\" err:%q", uri, err)
	}
	s, err := mgo.DialWithInfo(info)
	if err != nil {
		return nil, err
	}
	s.SetMode(mgo.Monotonic, true)
	return NewSession(s), nil
}

// C creates a new collection with the given name.
func (s *Session) C(db, coll string) (*Collection, error) {
	// Creating an index to speed up timeseries queries.
	index := mgo.Index{
		Key:        []string{typeField, timestampIndexField},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	c := s.DB(db).C(coll)
	if err := c.EnsureIndex(index); err != nil {
		return nil, err
	}
	return &Collection{c}, nil
}

// TSRecord represents a value to be added to timeseries database.
type TSRecord struct {
	Timestamp time.Time   `bson:"timestamp,omitempty"`  // bson name should match timestampIndexField.
	Value     interface{} `bson:"value,omitempty"`
}

// Collection represents a timeseries collection in a mongo database.
type Collection struct {
	*mgo.Collection
}

// TSUpsertResult holds the results for a timeseries upsert operation.
type TSUpsertResult struct {
	Matched  int
	Modified int
}

// TSUpsert bulk-inserts the given data into the timeseries database overriding the data if necessary.
func (c *Collection) TSUpsert(field string, val ...TSRecord) (TSUpsertResult, error) {
	switch len(val) {
	case 0:
		return TSUpsertResult{}, nil
	default:
		bulk := c.Bulk()
		for _, v := range val {
			bulk.Upsert(
				bson.M{timestampIndexField: v.Timestamp, typeField: field},
				bson.M{
					timestampIndexField: v.Timestamp,
					typeField:           field,
					valueField:          v.Value,
				},
			)
		}
		br, err := bulk.Run()
		return TSUpsertResult{br.Matched, br.Modified}, err
	}
}

// Interval fetches all records from timeseries mongo within the specified (closed) interval.
// If no records are found, an empty slice is returned.
func (c *Collection) Interval(field string, start time.Time, finish time.Time) ([]TSRecord, error) {
	iter := c.Find(
		bson.M{
			typeField: field,
			timestampIndexField: bson.M{
				"$gte": start,
				"$lte": finish,
			},
		}).Iter()

	var ret []TSRecord
	var d TSRecord
	for iter.Next(&d) {
		ret = append(ret, d)
	}
	if err := iter.Close(); err != nil {
		if err == mgo.ErrNotFound {
			return ret, nil
		}
		return nil, fmt.Errorf("Error querying tsmongo within range(%v,%v): %q", start, finish, err)
	}
	return ret, nil
}

// Last returns the last element in the timeseries, if any.
func (c *Collection) Last(field string) (TSRecord, error) {
	var r TSRecord
	err := c.Find(bson.M{typeField: field}).Sort("-" + timestampIndexField).One(&r)
	if err != nil {
		return TSRecord{}, err
	}
	return r, nil
}
