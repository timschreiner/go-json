package json_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/goccy/go-json"
)

type queryTestX struct {
	XA int
	XB string
	XC *queryTestY
	XD bool
	XE float32
}

type queryTestY struct {
	YA int
	YB string
	YC *queryTestZ
	YD bool
	YE float32
}

type queryTestZ struct {
	ZA string
	ZB bool
	ZC int
}

func TestFieldQuery(t *testing.T) {
	query, err := json.FieldQuery(
		"XA",
		"XB",
		json.SubFieldQuery("XC").Fields(
			"YA",
			"YB",
			json.SubFieldQuery("YC").Fields(
				"ZA",
				"ZB",
			),
		),
	).Build()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(query, &json.FieldQueryDef{
		Name: "",
		Fields: []*json.FieldQueryDef{
			{
				Name: "XA",
			},
			{
				Name: "XB",
			},
			{
				Name: "XC",
				Fields: []*json.FieldQueryDef{
					{
						Name: "YA",
					},
					{
						Name: "YB",
					},
					{
						Name: "YC",
						Fields: []*json.FieldQueryDef{
							{
								Name: "ZA",
							},
							{
								Name: "ZB",
							},
						},
					},
				},
			},
		},
	}) {
		t.Fatal("cannot get query")
	}
	queryStr, err := query.QueryString()
	if err != nil {
		t.Fatal(err)
	}
	if queryStr != `["XA","XB",{"XC":["YA","YB",{"YC":["ZA","ZB"]}]}]` {
		t.Fatalf("failed to create query string. %s", queryStr)
	}
	ctx := json.AddFieldQueryToContext(context.Background(), query)
	b, err := json.MarshalContext(ctx, &queryTestX{
		XA: 1,
		XB: "xb",
		XC: &queryTestY{
			YA: 2,
			YB: "yb",
			YC: &queryTestZ{
				ZA: "za",
				ZB: true,
				ZC: 3,
			},
			YD: true,
			YE: 4,
		},
		XD: true,
		XE: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}
