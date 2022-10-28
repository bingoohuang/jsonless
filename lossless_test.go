package jsonless

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Person struct {
	JSON `json:"-"`

	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string
	// JSON 格式化 time.Time 成 2006-01-02T15:04:05.999999999Z07:00 格式
	CreatedAt time.Time
	Ignored   bool `json:"-"`
}

func (p *Person) UnmarshalJSON(data []byte) error {
	return p.JSON.UnmarshalJSON(p, data)
}

func (p Person) MarshalJSON() ([]byte, error) {
	return p.JSON.MarshalJSON(p)
}

var jsondata = []byte(`
{"name": "Jack Wolfington",
 "age": 42,
 "address": "123 Fake St.",
 "CreatedAt": "2016-06-30T16:09:51.692226358+08:00",
 "Ignored": true,
 "Extra": {"foo": "bar"}}`)

func TestDecode(t *testing.T) {
	var p Person
	err := json.Unmarshal(jsondata, &p)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, p.Name, "Jack Wolfington")
	assert.Equal(t, p.Age, 42)
	assert.Equal(t, p.Address, "123 Fake St.")
	assert.Equal(t, p.CreatedAt, time.Date(2016, 6, 30, 16, 9, 51, 692226358, time.Local))
	assert.Equal(t, p.Ignored, false)
}

func TestEncode(t *testing.T) {
	now := time.Now()

	p := Person{
		Name:      "Wolf Jackington",
		Age:       33,
		Address:   "742 Evergreen Terrace",
		CreatedAt: now,
		Ignored:   true,
	}

	p.Set("Pi", 3.14159)

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := m["name"]
	assert.Equal(t, v, p.Name)

	v, ok = m["age"]
	assert.Equal(t, int(v.(float64)), p.Age)

	v, ok = m["Address"]
	assert.Equal(t, v, p.Address)

	v, ok = m["CreatedAt"]
	assert.Equal(t, v, p.CreatedAt.Format(time.RFC3339Nano))

	v, ok = m["Ignored"]
	assert.Equal(t, ok, false)

	v, ok = m["Pi"]
	assert.Equal(t, v, 3.14159)
}

func testDecodeEncode(t *testing.T) {
	var p Person
	err := json.Unmarshal(jsondata, &p)
	if err != nil {
		t.Fatal(err)
	}

	// Happy birthday, Jack
	p.Age++

	p.Ignored = true
	p.Set("age_printed", "forty-three")

	data, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := m["age"]
	assert.Equal(t, int(v.(float64)), p.Age)

	v, ok = m["Ignored"]
	assert.Equal(t, v, false)

	v, ok = m["Extra"]
	assert.Equal(t, ok, true)

	m2 := v.(map[string]interface{})
	v, ok = m2["foo"]
	assert.Equal(t, v, "bar")

	v, ok = m["age_printed"]
	assert.Equal(t, v, "forty-three")
}
