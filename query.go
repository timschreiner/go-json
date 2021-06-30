package json

import (
	"context"
	"fmt"
	"reflect"
)

type FieldQueryDef struct {
	Name   string
	Fields []*FieldQueryDef
}

func (d *FieldQueryDef) MarshalJSON() ([]byte, error) {
	if d.Name != "" {
		if len(d.Fields) > 0 {
			return Marshal(map[string][]*FieldQueryDef{d.Name: d.Fields})
		}
		return Marshal(d.Name)
	}
	return Marshal(d.Fields)
}

func (d *FieldQueryDef) QueryString() (FieldQueryString, error) {
	b, err := Marshal(d)
	if err != nil {
		return "", err
	}
	return FieldQueryString(b), nil
}

type FieldQueryString string

func (s FieldQueryString) Build() (*FieldQueryDef, error) {
	var query interface{}
	if err := Unmarshal([]byte(s), &query); err != nil {
		return nil, err
	}
	return s.build(reflect.ValueOf(query))
}

func (s FieldQueryString) build(v reflect.Value) (*FieldQueryDef, error) {
	switch v.Type().Kind() {
	case reflect.String:
		return s.buildString(v)
	case reflect.Map:
		return s.buildMap(v)
	case reflect.Slice:
		return s.buildSlice(v)
	case reflect.Interface:
		return s.build(reflect.ValueOf(v.Interface()))
	}
	return nil, fmt.Errorf("failed to build field query")
}

func (s FieldQueryString) buildString(v reflect.Value) (*FieldQueryDef, error) {
	b := []byte(v.String())
	switch b[0] {
	case '[', '{':
		var query interface{}
		if err := Unmarshal(b, &query); err != nil {
			return nil, err
		}
		if str, ok := query.(string); ok {
			return &FieldQueryDef{Name: str}, nil
		}
		return s.build(reflect.ValueOf(query))
	}
	return &FieldQueryDef{Name: string(b)}, nil
}

func (s FieldQueryString) buildSlice(v reflect.Value) (*FieldQueryDef, error) {
	fields := make([]*FieldQueryDef, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		def, err := s.build(v.Index(i))
		if err != nil {
			return nil, err
		}
		fields = append(fields, def)
	}
	return &FieldQueryDef{Fields: fields}, nil
}

func (s FieldQueryString) buildMap(v reflect.Value) (*FieldQueryDef, error) {
	keys := v.MapKeys()
	if len(keys) != 1 {
		return nil, fmt.Errorf("failed to build field query object")
	}
	key := keys[0]
	if key.Type().Kind() != reflect.String {
		return nil, fmt.Errorf("failed to build field query. invalid object key type")
	}
	name := key.String()
	def, err := s.build(v.MapIndex(key))
	if err != nil {
		return nil, err
	}
	return &FieldQueryDef{
		Name:   name,
		Fields: def.Fields,
	}, nil
}

func FieldQuery(fields ...FieldQueryString) FieldQueryString {
	query, _ := Marshal(fields)
	return FieldQueryString(query)
}

func SubFieldQuery(field string) *SubFieldQueryDef {
	return &SubFieldQueryDef{field: field}
}

type SubFieldQueryDef struct {
	field string
}

func (q *SubFieldQueryDef) Fields(fields ...FieldQueryString) FieldQueryString {
	query, _ := Marshal(map[string][]FieldQueryString{q.field: fields})
	return FieldQueryString(query)
}

type queryKey struct{}

func FieldQueryFromContext(ctx context.Context) *FieldQueryDef {
	query := ctx.Value(queryKey{})
	if query == nil {
		return nil
	}
	q, ok := query.(*FieldQueryDef)
	if !ok {
		return nil
	}
	return q
}

func AddFieldQueryToContext(ctx context.Context, query *FieldQueryDef) context.Context {
	return context.WithValue(ctx, queryKey{}, query)
}
