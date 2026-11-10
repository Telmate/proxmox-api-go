package proxmox

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type testRaw struct {
	Port   *string `json:"port"`
	Ticket *string `json:"ticket"`
	ID     int     `json:"id"`
	Name   *string `json:"name"`
	User   *string `json:"user"`
}

type testFinal struct {
	Port   *int
	Ticket *string
	ID     *GuestID
	Name   *GuestName
	User   *UserID
}

func testData_Json() []struct {
	name   string
	f      func([]byte) (*testFinal, error)
	input  []byte
	output *testFinal
	err    bool
} {
	out := testFinal{
		Port:   new(9050),
		Ticket: new("ccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"),
		ID:     new(GuestID(100)),
		Name:   new(GuestName("B.Wayne")),
		User:   &UserID{Name: "root", Realm: "pam"},
	}

	raw, _ := json.Marshal(responseData[testRaw]{Data: &testRaw{
		Port:   new(strconv.Itoa(*out.Port)),
		Ticket: out.Ticket,
		ID:     100,
		User:   new("root@pam"),
		Name:   new(out.Name.String()),
	}})

	parse := func(vv map[string]any) (*testFinal, error) {
		var t testFinal
		if v, ok := vv["name"]; ok {
			t.Name = new(GuestName(v.(string)))
		}
		if v, ok := vv["ticket"]; ok {
			t.Ticket = new(v.(string))
		}
		if v, ok := vv["port"]; ok {
			tmp, err := strconv.ParseInt(v.(string), 10, 64)
			if err == nil {
				t.Port = new(int(tmp))
			}
		}
		if v, ok := vv["id"]; ok {
			t.ID = new(GuestID(v.(float64)))
		}
		if v, ok := vv["user"]; ok {
			var tmp UserID
			if err := tmp.Parse(v.(string)); err == nil {
				t.User = &tmp
			}
		}
		return &t, nil
	}

	transform := func(raw *testRaw) *testFinal {
		if raw == nil {
			return nil
		}
		tmp := testFinal{
			ID: new(GuestID(raw.ID)),
		}
		if raw.Name != nil {
			tmp.Name = new(GuestName(*raw.Name))
		}
		if raw.Port != nil {
			v, err := strconv.ParseInt(*raw.Port, 10, 64)
			if err == nil {
				tmp.Port = new(int(v))
			}
		}
		if raw.Ticket != nil {
			tmp.Ticket = raw.Ticket
		}
		if raw.User != nil {
			var v UserID
			if err := v.Parse(*raw.User); err == nil {
				tmp.User = &v
			}
		}
		return &tmp
	}

	return []struct {
		name   string
		f      func([]byte) (*testFinal, error)
		input  []byte
		output *testFinal
		err    bool
	}{
		{name: `any`,
			input:  raw,
			output: &out,
			f: func(b []byte) (*testFinal, error) {
				var data any
				json.Unmarshal(b, &data)
				if v, ok := data.(map[string]any)["data"]; ok {
					return parse(v.(map[string]any))
				}
				return nil, nil
			}},
		{name: `struct`,
			input:  raw,
			output: &out,
			f: func(b []byte) (*testFinal, error) {
				var data responseData[testRaw]
				err := json.Unmarshal(b, &data)
				return transform(data.Data), err
			}},
		{name: `mixed`,
			input:  raw,
			output: &out,
			f: func(b []byte) (*testFinal, error) {
				var data responseData[map[string]any]
				if err := json.Unmarshal(b, &data); err != nil {
					return nil, err
				}
				return parse(*data.Data)
			}},
		{name: `error struct`,
			input:  []byte(`{"meaning":42}`),
			output: nil,
			err:    true,
			f: func(b []byte) (*testFinal, error) {
				var data responseData[testRaw]
				err := json.Unmarshal(b, &data)
				return transform(data.Data), err
			}},
	}
}

func Test_Json(t *testing.T) {
	t.Parallel()
	tests := testData_Json()
	for _, test := range tests {
		t.Run(test.name, func(*testing.T) {
			out, err := test.f(test.input)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.output, out)
		})
	}
}

func Benchmark_Json(b *testing.B) {
	tests := testData_Json()
	for i := range tests {
		b.Run(tests[i].name, func(b *testing.B) {
			b.ResetTimer()
			for b.Loop() {
				tests[i].f(tests[i].input)
			}
		})
	}
}
