# jsonless

jsonless is a Go library that populates structs from JSON and
allows serialization back to JSON without losing fields that are
not explicitly defined in the struct.

Forked from [joeshaw/json-lossless](https://github.com/joeshaw/json-lossless).

## API

To get started, embed a `jsonless.JSON` inside your struct:

```go
type Person struct {
	jsonless.JSON `json:"-"`

	Name      string `json:"name"`
	Age       int    `json:"age"`
	Address   string
	CreatedAt time.Time
}
```

Define `MarshalJSON` and `UnmarshalJSON` methods on the type
to implement the `json.Marshaler` and `json.Unmarshaler` interfaces,
deferring the work to the `jsonless.JSON` embed:

```go
func (p *Person) UnmarshalJSON(data []byte) error {
	return p.JSON.UnmarshalJSON(p, data)
}

func (p Person) MarshalJSON() ([]byte, error) {
	return p.JSON.MarshalJSON(p)
}
```

Given JSON like this:

```json
{
  "name": "Jack Wolfington",
  "age": 42,
  "address": "123 Fake St.",
  "CreatedAt": "2013-09-16T10:44:40.295451647-00:00",
  "Extra": {
    "foo": "bar"
  }
}
```

When you decode into a struct, the `Extra` field will be kept around,
even though it's not accessible from your struct.

```go
var p Person
if err := json.Unmarshal(data, &p); err != nil {
	panic(err)
}

data, err := json.Marshal(p)
if err != nil {
	panic(err)
}

// "Extra" is still set in the marshaled JSON:
if bytes.Index(data, "Extra") == -1 {
	panic("Extra not in data!")
}

fmt.Println(string(data))

```

You can also set arbitrary key/values on your struct by calling `Set()`:

```go
p.Set("Extra", "AgeString", "forty-two")
```

When serialized, `Extra` will look like this:

```json
{
  "...": "...",
  "Extra": {
    "foo": "bar",
    "AgeString": "forty-two"
  }
}
```
