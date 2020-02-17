package lookup

import (
	"fmt"
	"reflect"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestLookup_Map(c *C) {
	value, err := Lookup(map[string]int{"foo": 42}, false, "foo")
	c.Assert(err, IsNil)
	c.Assert(value.Int(), Equals, int64(42))
}

func (s *S) TestLookup_Ptr(c *C) {
	value, err := Lookup(&structFixture, false, "String")
	c.Assert(err, IsNil)
	c.Assert(value.String(), Equals, "foo")
}

func (s *S) TestLookup_Interface(c *C) {
	value, err := Lookup(structFixture, false, "Interface")

	c.Assert(err, IsNil)
	c.Assert(value.String(), Equals, "foo")
}

func (s *S) TestLookup_StructBasic(c *C) {
	value, err := Lookup(structFixture, false, "String")
	c.Assert(err, IsNil)
	c.Assert(value.String(), Equals, "foo")
}

func (s *S) TestLookup_StructPlusMap(c *C) {
	value, err := Lookup(structFixture, false, "Map", "foo")
	c.Assert(err, IsNil)
	c.Assert(value.Int(), Equals, int64(42))
}

func (s *S) TestLookup_MapNamed(c *C) {
	value, err := Lookup(mapFixtureNamed, false, "foo")
	c.Assert(err, IsNil)
	c.Assert(value.Int(), Equals, int64(42))
}

func (s *S) TestLookup_NotFound(c *C) {
	_, err := Lookup(structFixture, false, "qux")
	c.Assert(err, Equals, ErrKeyNotFound)

	_, err = Lookup(mapFixture, false, "qux")
	c.Assert(err, Equals, ErrKeyNotFound)
}

func (s *S) TestAggregableLookup_StructIndex(c *C) {
	value, err := Lookup(structFixture, false, "StructSlice", "Map", "foo")

	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, []int{42, 42})
}

func (s *S) TestAggregableLookup_StructNestedMap(c *C) {
	value, err := Lookup(structFixture, false, "StructSlice[0]", "String")

	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, "foo")
}

func (s *S) TestAggregableLookup_StructNested(c *C) {
	value, err := Lookup(structFixture, false, "StructSlice", "StructSlice", "String")

	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, []string{"bar", "foo", "qux", "baz"})
}

func (s *S) TestAggregableLookupString_Complex(c *C) {
	value, err := LookupString(structFixture, "StructSlice.StructSlice[0].String", false)
	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, []string{"bar", "foo", "qux", "baz"})

	value, err = LookupString(structFixture, "StructSlice[0].Map.foo", false)
	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, 42)

	value, err = LookupString(mapComplexFixture, "map.bar", false)
	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, 1)

	value, err = LookupString(mapComplexFixture, "list.baz", false)
	c.Assert(err, IsNil)
	c.Assert(value.Interface(), DeepEquals, []int{1, 2, 3})
}

func (s *S) TestAggregableLookup_EmptySlice(c *C) {
	fixture := [][]MyStruct{{}}
	value, err := LookupString(fixture, "String", false)
	c.Assert(err, IsNil)
	c.Assert(value.Interface().([]string), DeepEquals, []string{})
}

func (s *S) TestAggregableLookup_EmptyMap(c *C) {
	fixture := map[string]*MyStruct{}
	value, err := LookupString(fixture, "Map", false)
	c.Assert(err, IsNil)
	c.Assert(value.Interface().([]map[string]int), DeepEquals, []map[string]int{})
}

func (s *S) TestMergeValue(c *C) {
	v := mergeValue([]reflect.Value{reflect.ValueOf("qux"), reflect.ValueOf("foo")})
	c.Assert(v.Interface(), DeepEquals, []string{"qux", "foo"})
}

func (s *S) TestMergeValueSlice(c *C) {
	v := mergeValue([]reflect.Value{
		reflect.ValueOf([]string{"foo", "bar"}),
		reflect.ValueOf([]string{"qux", "baz"}),
	})

	c.Assert(v.Interface(), DeepEquals, []string{"foo", "bar", "qux", "baz"})
}

func (s *S) TestMergeValueZero(c *C) {
	v := mergeValue([]reflect.Value{reflect.Value{}, reflect.ValueOf("foo")})
	c.Assert(v.Interface(), DeepEquals, []string{"foo"})
}

func (s *S) TestParseIndex(c *C) {
	key, index, err := parseIndex("foo[42]")
	c.Assert(err, IsNil)
	c.Assert(key, Equals, "foo")
	c.Assert(index, Equals, 42)
}

func (s *S) TestParseIndexNooIndex(c *C) {
	key, index, err := parseIndex("foo")
	c.Assert(err, IsNil)
	c.Assert(key, Equals, "foo")
	c.Assert(index, Equals, -1)
}

func (s *S) TestParseIndexMalFormed(c *C) {
	key, index, err := parseIndex("foo[]")
	c.Assert(err, Equals, ErrMalformedIndex)
	c.Assert(key, Equals, "")
	c.Assert(index, Equals, -1)

	key, index, err = parseIndex("foo[42")
	c.Assert(err, Equals, ErrMalformedIndex)
	c.Assert(key, Equals, "")
	c.Assert(index, Equals, -1)

	key, index, err = parseIndex("foo42]")
	c.Assert(err, Equals, ErrMalformedIndex)
	c.Assert(key, Equals, "")
	c.Assert(index, Equals, -1)
}

type User struct {
	Name    string  `json:"name_field"`
	Address Address `json:"address"`
}

type Address struct {
	FirstLine  string      `json:"first_line"`
	SecondLine string      `json:"second_line"`
	PostalCode *PostalCode `json:"postal_code"`
}

type PostalCode struct {
	zip   int    `json:"zip"`
	state string `json:"state"`
}

func (s *S) TestJSON(c *C) {
	user := &User{
		Name: "John Doe",
		Address: Address{
			FirstLine:  "High",
			SecondLine: "Street",
			PostalCode: &PostalCode{
				zip:   1234,
				state: "NSW",
			},
		},
	}
	val, err := LookupString(user, "address.first_line", true)
	c.Assert(err, IsNil)
	c.Assert(val.String(), Equals, "High")

	val, err = LookupString(user, "address.postal_code", true)
	c.Assert(err, IsNil)
	c.Assert(val.Interface().(*PostalCode).state, Equals, "NSW")

	val, err = LookupString(user, "address.postal_code.zip", true)
	c.Assert(err, IsNil)
	c.Assert(val.Int(), Equals, int64(1234))
}

func getStructTag(f reflect.StructField) string {
	return string(f.Tag)
}

func ExampleLookupString() {
	type Cast struct {
		Actor, Role string
	}

	type Serie struct {
		Cast []Cast
	}

	series := map[string]Serie{
		"A-Team": {Cast: []Cast{
			{Actor: "George Peppard", Role: "Hannibal"},
			{Actor: "Dwight Schultz", Role: "Murdock"},
			{Actor: "Mr. T", Role: "Baracus"},
			{Actor: "Dirk Benedict", Role: "Faceman"},
		}},
	}

	q := "A-Team.Cast.Role"
	value, _ := LookupString(series, q, false)
	fmt.Println(q, "->", value.Interface())

	q = "A-Team.Cast[0].Actor"
	value, _ = LookupString(series, q, false)
	fmt.Println(q, "->", value.Interface())

	// Output:
	// A-Team.Cast.Role -> [Hannibal Murdock Baracus Faceman]
	// A-Team.Cast[0].Actor -> George Peppard
}

func ExampleLookup() {
	type ExampleStruct struct {
		Values struct {
			Foo int
		}
	}

	i := ExampleStruct{}
	i.Values.Foo = 10

	value, _ := Lookup(i, false, "Values", "Foo")
	fmt.Println(value.Interface())
	// Output: 10
}

type MyStruct struct {
	String      string
	Map         map[string]int
	Nested      *MyStruct
	StructSlice []*MyStruct
	Interface   interface{}
}

type MyKey string

var mapFixtureNamed = map[MyKey]int{"foo": 42}
var mapFixture = map[string]int{"foo": 42}
var structFixture = MyStruct{
	String:    "foo",
	Map:       mapFixture,
	Interface: "foo",
	StructSlice: []*MyStruct{
		{Map: mapFixture, String: "foo", StructSlice: []*MyStruct{{String: "bar"}, {String: "foo"}}},
		{Map: mapFixture, String: "qux", StructSlice: []*MyStruct{{String: "qux"}, {String: "baz"}}},
	},
}

var mapComplexFixture = map[string]interface{}{
	"map": map[string]interface{}{
		"bar": 1,
	},
	"list": []map[string]interface{}{
		{"baz": 1},
		{"baz": 2},
		{"baz": 3},
	},
}
