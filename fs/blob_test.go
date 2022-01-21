package fs

import (
	"testing"
)

/*
type uuid [16]byte

type key []byte

func Test_codec(t *testing.T) {
	t.Log("->", must(encode(int32(1))))
	t.Log("->", must(encode(int32(255))))
	t.Log("->", must(encode(int32(-127))))
	t.Log("->", must(encode(uuid{})))
	t.Log("->", must(encode("550e8400-e29b-11d4-a716-446655440000")))
	t.Log("->", must(encode("ticket/550e8400-e29b-11d4-a716-446655440000/attachments/screenshot.jpg")))
	t.Log("->", must(encode([]byte("%&$§\"öäü@!:;/\\"))))
	/*
		tests := []struct {
			id any
		}{
			//{id: nil},
			{id: ""},
			{id: float64(1)},
			{id: float64(10)},
			{id: float64(100)},
			{id: "abcdefghijklmnoprstvwxyz1234567890"},
			{id: "%&$§\"öäü@!:;/\\"},
			{id: "550e8400-e29b-11d4-a716-446655440000"},
			{id: uuid{}},
			{id: key{}},
		}
		for i, tt := range tests {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				path, err := encode[any](tt.id)
				if err != nil {
					t.Fatal(err)
				}
				t.Log(tt.id, "=>", path, len(fmt.Sprintf("%v", tt.id)), ":", len(path)-(3+4))
				dec, err := decode[any](path)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(dec, tt.id) {
					t.Errorf("decode() got = %v (%v), want %v (%v)", dec, reflect.TypeOf(dec), tt.id, reflect.TypeOf(tt.id))
				}
			})
		}*/
//}*/

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

func Test_validName(t *testing.T) {

	tests := []struct {
		name string
		want bool
	}{
		{"", false},
		{"/", false},
		{".", false},
		{"a/b/c", true},
		{"a/b /c", false},
		{"a/b b/c", false},
		{"A/b/c", false},
		{"%20/b/c", false},
		{"/a", false},
		{"a", true},
		{"a/b/./c", false},
		{"a/b//c", false},
		{"a/b/../c", false},
		{"a/b/c/", false},
		{"myname.txt", true},
		{"my-context/my_ticket/file.txt", true},
		{"1/2/3", true},
		{"550e8400-e29b-11d4-a716-446655440000", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidName(tt.name); got != tt.want {
				t.Errorf("validName() = %v, want %v", got, tt.want)
			}
		})
	}
}
