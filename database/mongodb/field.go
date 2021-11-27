package mongodb

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"reflect"
	"time"
)

type CommonField struct {
	Initial     bool               `bson:"initial"`
	Description string             `bson:"description"`
	CreatedAt   primitive.DateTime `bson:"created_at"`
	UpdatedAt   primitive.DateTime `bson:"updated_at"`
}

// See
func (f CommonField) UnmarshalBSON([]byte) error {

	return nil
}

func (f CommonField) GetBSON() (interface{}, error) {
	return nil, nil

}

func (f CommonField) SetBSON(raw bson.RawValue) error {
	return nil

}

type timeCodec struct {
	bsoncodec.TimeCodec
}

func (tc *timeCodec) EncodeValue(ec bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	if val.IsZero() {
		// Fix unaddressable value
		value := reflect.New(val.Type()).Elem()
		value.Set(reflect.ValueOf(time.Now().UTC()))
		val = value
	}
	return tc.TimeCodec.EncodeValue(ec, vw, val)
}
