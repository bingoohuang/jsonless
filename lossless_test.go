package jsonless

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/bingoohuang/gg/pkg/mapstruct"
	"github.com/stretchr/testify/assert"
)

type Address struct {
	Detail   string `json:"detail"`
	PostCode string `json:"postCode"`
}

type Person struct {
	JSON `json:"-"`

	Name    string  `json:"name"`
	Age     int     `json:"age"`
	Address Address `json:"address"`
	Words   []string
	// JSON 格式化 time.Time 成 2006-01-02T15:04:05.999999999Z07:00 格式
	CreatedAt time.Time
	Ignored   bool   `json:"-"`
	Omit      string `json:"omit,omitempty"`
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
 "address": { "detail": "123 Fake St.", "postCode":"123456"},
 "words": ["aa","bb"],
 "CreatedAt": "2016-06-30T16:09:51.692226358+08:00",
 "Ignored": true,
 "Extra": {"foo": "bar"}}`)

func TestDecode(t *testing.T) {
	var p Person
	err := json.Unmarshal(jsondata, &p)
	assert.Nil(t, err)

	assert.Equal(t, p.Name, "Jack Wolfington")
	assert.Equal(t, p.Age, 42)
	assert.Equal(t, []string{"aa", "bb"}, p.Words)
	assert.Equal(t, Address{Detail: "123 Fake St.", PostCode: "123456"}, p.Address)
	assert.Equal(t, p.CreatedAt, time.Date(2016, 6, 30, 16, 9, 51, 692226358, time.Local))
	assert.Equal(t, p.Ignored, false)
}

func TestEncode(t *testing.T) {
	now := time.Now()

	p := Person{
		Name:      "Wolf Jackington",
		Age:       33,
		Address:   Address{Detail: "742 Evergreen Terrace", PostCode: "123123"},
		CreatedAt: now,
		Ignored:   true,
	}

	p.Set("Pi", 3.14159)
	p.Set("omit", "anything")

	data, err := json.Marshal(p)
	assert.Nil(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	assert.Nil(t, err)

	v, ok := m["name"]
	assert.Equal(t, v, p.Name)

	v, ok = m["age"]
	assert.Equal(t, int(v.(float64)), p.Age)

	v, ok = m["address"]
	var addr Address
	mapstruct.Decode(v, &addr)
	assert.Equal(t, addr, p.Address)

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
	assert.Nil(t, err)

	// Happy birthday, Jack
	p.Age++

	p.Ignored = true
	p.Set("age_printed", "forty-three")

	data, err := json.Marshal(p)
	assert.Nil(t, err)

	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	assert.Nil(t, err)

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
