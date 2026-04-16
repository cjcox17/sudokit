package events

import "go.mongodb.org/mongo-driver/bson/primitive"

type FieldChange[T any] struct {
	Old T `json:"old"`
	New T `json:"new"`
}

type OptionalFieldChange[T any] struct {
	Old *T `json:"old,omitempty"`
	New *T `json:"new,omitempty"`
}

type ObjectIDChange struct {
	Old primitive.ObjectID `json:"old"`
	New primitive.ObjectID `json:"new"`
}

type OptionalObjectIDChange struct {
	Old *primitive.ObjectID `json:"old,omitempty"`
	New *primitive.ObjectID `json:"new,omitempty"`
}

type StringChange = FieldChange[string]
type BoolChange = FieldChange[bool]
type IntChange = FieldChange[int]
type Int64Change = FieldChange[int64]
type Float64Change = FieldChange[float64]
