package main

import (
	"encoding/json"
	_ "encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	. "github.com/smartystreets/goconvey/convey"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/agiledragon/gomonkey/v2/test/fake"
)

func TestMyModBase(t *testing.T) {
	var (
		a    = 7
		b    = 7
		want = 0
	)

	actual := MyMod(a, b)
	if actual != want {
		t.Errorf("MyMod() = %v, want %v", actual, want)
	}
}

func TestMyModBaseAssert(t *testing.T) {
	var (
		a    = 7
		b    = 7
		want = 0
	)

	assert.Equal(t, MyMod(a, b), want)
}

func TestMyMod(t *testing.T) {
	tests := []struct {
		A    int
		B    int
		Want int
		Desc string
	}{
		{1, 2, 1, "a < b"},
		{31, 9, 4, "a > b"},
		{10, 0, -1, " chushu is 0"},
		{8, 2, 0, "刚好整除"},
	}
	for _, tt := range tests {
		t.Run(tt.Desc, func(t *testing.T) {
			if got := MyMod(tt.A, tt.B); got != tt.Want {
				t.Errorf("MyMod() = %v, want %v", got, tt.Want)
			}
		})
	}
}

func TestMyModAssert(t *testing.T) {
	tests := []struct {
		A    int
		B    int
		Want int
		Desc string
	}{
		{1, 2, 1, "a < b"},
		{31, 9, 4, "a > b"},
		{10, 0, -1, " chushu is 0"},
		{8, 2, 0, "刚好整除"},
	}
	for _, tt := range tests {
		assert.Equal(t, MyMod(tt.A, tt.B), tt.Want, tt.Desc)
	}
}

func BenchmarkMyMod(b *testing.B) {
	for n := 0; n < b.N; n++ {
		a := 3
		b := 4
		MyMod(a, b)

	}
}

func ExampleHello() {
	fmt.Println("Hello")
	// Output: Hello
}

var db struct {
	Dns string
}

func TestMain(m *testing.M) {
	db.Dns = os.Getenv("DATABASE_DNS")
	if db.Dns == "" {
		db.Dns = "root:123456@tcp(localhost:3306)/?charset=utf8&parseTime=True&loc=Local"
	}

	flag.Parse()
	exitCode := m.Run()

	db.Dns = ""

	// 退出
	os.Exit(exitCode)
}

func TestDatabase(t *testing.T) {
	fmt.Println(db.Dns)
}

var (
	outputExpect = "xxx-vethName100-yyy"
)

func TestMymodFunc(t *testing.T) {
	flag.Parse()
	Convey("TestMymodFunc", t, func() {
		patches := ApplyFunc(MyMod, func(_ int, _ int) int {
			return 1
		})
		defer patches.Reset()

		var (
			a    = 7
			b    = 7
			want = 0
		)

		So(MyMod(a, b), ShouldEqual, want)
	})
}

// stub fun
func TestApplyFunc(t *testing.T) {
	Convey("TestApplyFunc", t, func() {

		Convey("one func for succ", func() {
			patches := ApplyFunc(fake.Exec, func(_ string, _ ...string) (string, error) {
				return outputExpect, nil
			})
			defer patches.Reset()
			output, err := fake.Exec("", "")
			// So(err, ShouldEqual, nil)
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, outputExpect)
		})

		Convey("one func for fail", func() {
			patches := ApplyFunc(fake.Exec, func(_ string, _ ...string) (string, error) {
				return "", fake.ErrActual
			})
			defer patches.Reset()
			output, err := fake.Exec("", "")
			So(err, ShouldEqual, fake.ErrActual)
			So(output, ShouldEqual, "")
		})

		Convey("two funcs", func() {
			patches := ApplyFunc(fake.Exec, func(_ string, _ ...string) (string, error) {
				return outputExpect, nil
			})
			defer patches.Reset()
			patches.ApplyFunc(fake.Belong, func(_ string, _ []string) bool {
				return true
			})
			output, err := fake.Exec("", "")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, outputExpect)
			flag := fake.Belong("", nil)
			So(flag, ShouldBeTrue)
		})

		Convey("input and output param", func() {
			patches := ApplyFunc(json.Unmarshal, func(_ []byte, v interface{}) error {
				p := v.(*map[int]int)
				*p = make(map[int]int)
				(*p)[1] = 2
				(*p)[2] = 4
				return nil
			})
			defer patches.Reset()
			var m map[int]int
			err := json.Unmarshal([]byte("123"), &m)
			So(err, ShouldEqual, nil)
			So(m[1], ShouldEqual, 2)
			So(m[2], ShouldEqual, 4)
		})
	})
}

func TestApplyMethod(t *testing.T) {
	slice := fake.NewSlice()
	var s *fake.Slice
	Convey("TestApplyMethod", t, func() {

		Convey("for succ", func() {
			err := slice.Add(1)
			So(err, ShouldEqual, nil)
			patches := ApplyMethod(reflect.TypeOf(s), "Add", func(s1 *fake.Slice, i int) error {
				return nil
			})

			fmt.Println(slice)
			defer patches.Reset()
			err = slice.Add(1)
			fmt.Println(slice)
			So(err, ShouldEqual, nil)
			err = slice.Remove(1)
			So(err, ShouldEqual, nil)
			So(len(slice), ShouldEqual, 0)
		})

		Convey("for already exist", func() {
			err := slice.Add(2)
			So(err, ShouldEqual, nil)
			patches := ApplyMethod(reflect.TypeOf(s), "Add", func(_ *fake.Slice, _ int) error {
				return fake.ErrElemExsit
			})
			defer patches.Reset()
			err = slice.Add(1)
			So(err, ShouldEqual, fake.ErrElemExsit)
			err = slice.Remove(2)
			So(err, ShouldEqual, nil)
			So(len(slice), ShouldEqual, 0)
		})

		Convey("two methods", func() {
			err := slice.Add(3)
			So(err, ShouldEqual, nil)
			defer slice.Remove(3)
			patches := ApplyMethod(reflect.TypeOf(s), "Add", func(_ *fake.Slice, _ int) error {
				return fake.ErrElemExsit
			})
			defer patches.Reset()
			patches.ApplyMethod(reflect.TypeOf(s), "Remove", func(_ *fake.Slice, _ int) error {
				return fake.ErrElemNotExsit
			})
			err = slice.Add(2)
			So(err, ShouldEqual, fake.ErrElemExsit)
			err = slice.Remove(1)
			So(err, ShouldEqual, fake.ErrElemNotExsit)
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldEqual, 3)
		})

		Convey("one func and one method", func() {
			err := slice.Add(4)
			So(err, ShouldEqual, nil)
			defer slice.Remove(4)
			patches := ApplyFunc(fake.Exec, func(_ string, _ ...string) (string, error) {
				return outputExpect, nil
			})
			defer patches.Reset()
			patches.ApplyMethod(reflect.TypeOf(s), "Remove", func(_ *fake.Slice, _ int) error {
				return fake.ErrElemNotExsit
			})
			output, err := fake.Exec("", "")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, outputExpect)
			err = slice.Remove(1)
			So(err, ShouldEqual, fake.ErrElemNotExsit)
			So(len(slice), ShouldEqual, 1)
			So(slice[0], ShouldEqual, 4)
		})
	})
}

var num = 10

func TestApplyGlobalVar(t *testing.T) {
	Convey("TestApplyGlobalVar", t, func() {

		Convey("change", func() {
			patches := ApplyGlobalVar(&num, 150)
			defer patches.Reset()
			So(num, ShouldEqual, 150)
		})

		Convey("recover", func() {
			So(num, ShouldEqual, 10)
		})
	})
}

func TestApplyFuncVar(t *testing.T) {
	Convey("TestApplyFuncVar", t, func() {

		Convey("for succ", func() {
			str := "hello"
			patches := ApplyFuncVar(&fake.Marshal, func(_ interface{}) ([]byte, error) {
				return []byte(str), nil
			})
			defer patches.Reset()
			bytes, err := fake.Marshal(nil)
			So(err, ShouldEqual, nil)
			So(string(bytes), ShouldEqual, str)
		})

		Convey("for fail", func() {
			patches := ApplyFuncVar(&fake.Marshal, func(_ interface{}) ([]byte, error) {
				return nil, fake.ErrActual
			})
			defer patches.Reset()
			_, err := fake.Marshal(nil)
			So(err, ShouldEqual, fake.ErrActual)
		})
	})
}

func TestApplyFuncSeq(t *testing.T) {
	Convey("TestApplyFuncSeq", t, func() {

		Convey("default times is 1", func() {
			info1 := "hello cpp"
			info2 := "hello golang"
			info3 := "hello gomonkey"
			outputs := []OutputCell{
				{Values: Params{info1, nil}},
				{Values: Params{info2, nil}},
				{Values: Params{info3, nil}},
			}
			patches := ApplyFuncSeq(fake.ReadLeaf, outputs)
			defer patches.Reset()

			runtime.GC()

			output, err := fake.ReadLeaf("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
			output, err = fake.ReadLeaf("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info2)
			output, err = fake.ReadLeaf("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info3)
		})

		Convey("retry succ util the third times", func() {
			info1 := "hello cpp"
			outputs := []OutputCell{
				{Values: Params{"", fake.ErrActual}, Times: 2},
				{Values: Params{info1, nil}},
			}
			patches := ApplyFuncSeq(fake.ReadLeaf, outputs)
			defer patches.Reset()
			output, err := fake.ReadLeaf("")
			So(err, ShouldEqual, fake.ErrActual)
			output, err = fake.ReadLeaf("")
			So(err, ShouldEqual, fake.ErrActual)
			output, err = fake.ReadLeaf("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
		})

		Convey("batch operations failed on the third time", func() {
			info1 := "hello gomonkey"
			outputs := []OutputCell{
				{Values: Params{info1, nil}, Times: 2},
				{Values: Params{"", fake.ErrActual}},
			}
			patches := ApplyFuncSeq(fake.ReadLeaf, outputs)
			defer patches.Reset()
			output, err := fake.ReadLeaf("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
			output, err = fake.ReadLeaf("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
			output, err = fake.ReadLeaf("")
			So(err, ShouldEqual, fake.ErrActual)
		})

	})
}

func TestApplyMethodSeq(t *testing.T) {
	e := &fake.Etcd{}
	Convey("TestApplyMethodSeq", t, func() {

		Convey("default times is 1", func() {
			info1 := "hello cpp"
			info2 := "hello golang"
			info3 := "hello gomonkey"
			outputs := []OutputCell{
				{Values: Params{info1, nil}},
				{Values: Params{info2, nil}},
				{Values: Params{info3, nil}},
			}
			patches := ApplyMethodSeq(reflect.TypeOf(e), "Retrieve", outputs)
			defer patches.Reset()
			output, err := e.Retrieve("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
			output, err = e.Retrieve("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info2)
			output, err = e.Retrieve("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info3)
		})

		Convey("retry succ util the third times", func() {
			info1 := "hello cpp"
			outputs := []OutputCell{
				{Values: Params{"", fake.ErrActual}, Times: 2},
				{Values: Params{info1, nil}},
			}
			patches := ApplyMethodSeq(reflect.TypeOf(e), "Retrieve", outputs)
			defer patches.Reset()
			output, err := e.Retrieve("")
			So(err, ShouldEqual, fake.ErrActual)
			output, err = e.Retrieve("")
			So(err, ShouldEqual, fake.ErrActual)
			output, err = e.Retrieve("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
		})

		Convey("batch operations failed on the third time", func() {
			info1 := "hello gomonkey"
			outputs := []OutputCell{
				{Values: Params{info1, nil}, Times: 2},
				{Values: Params{"", fake.ErrActual}},
			}
			patches := ApplyMethodSeq(reflect.TypeOf(e), "Retrieve", outputs)
			defer patches.Reset()
			output, err := e.Retrieve("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
			output, err = e.Retrieve("")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, info1)
			output, err = e.Retrieve("")
			So(err, ShouldEqual, fake.ErrActual)
		})

	})
}

func TestApplyFuncVarSeq(t *testing.T) {
	Convey("TestApplyFuncVarSeq", t, func() {

		Convey("default times is 1", func() {
			info1 := "hello cpp"
			info2 := "hello golang"
			info3 := "hello gomonkey"
			outputs := []OutputCell{
				{Values: Params{[]byte(info1), nil}},
				{Values: Params{[]byte(info2), nil}},
				{Values: Params{[]byte(info3), nil}},
			}
			patches := ApplyFuncVarSeq(&fake.Marshal, outputs)
			defer patches.Reset()
			bytes, err := fake.Marshal("")
			So(err, ShouldEqual, nil)
			So(string(bytes), ShouldEqual, info1)
			bytes, err = fake.Marshal("")
			So(err, ShouldEqual, nil)
			So(string(bytes), ShouldEqual, info2)
			bytes, err = fake.Marshal("")
			So(err, ShouldEqual, nil)
			So(string(bytes), ShouldEqual, info3)
		})

		Convey("retry succ util the third times", func() {
			info1 := "hello cpp"
			outputs := []OutputCell{
				{Values: Params{[]byte(""), fake.ErrActual}, Times: 2},
				{Values: Params{[]byte(info1), nil}},
			}
			patches := ApplyFuncVarSeq(&fake.Marshal, outputs)
			defer patches.Reset()
			bytes, err := fake.Marshal("")
			So(err, ShouldEqual, fake.ErrActual)
			bytes, err = fake.Marshal("")
			So(err, ShouldEqual, fake.ErrActual)
			bytes, err = fake.Marshal("")
			So(err, ShouldEqual, nil)
			So(string(bytes), ShouldEqual, info1)
		})

		Convey("batch operations failed on the third time", func() {
			info1 := "hello gomonkey"
			outputs := []OutputCell{
				{Values: Params{[]byte(info1), nil}, Times: 2},
				{Values: Params{[]byte(""), fake.ErrActual}},
			}
			patches := ApplyFuncVarSeq(&fake.Marshal, outputs)
			defer patches.Reset()
			bytes, err := fake.Marshal("")
			So(err, ShouldEqual, nil)
			So(string(bytes), ShouldEqual, info1)
			bytes, err = fake.Marshal("")
			So(err, ShouldEqual, nil)
			So(string(bytes), ShouldEqual, info1)
			bytes, err = fake.Marshal("")
			So(err, ShouldEqual, fake.ErrActual)
		})

	})
}

func TestPatchPair(t *testing.T) {

	Convey("TestPatchPair", t, func() {

		Convey("TestPatchPair", func() {
			patchPairs := [][2]interface{}{
				{
					fake.Exec,
					func(_ string, _ ...string) (string, error) {
						return outputExpect, nil
					},
				},
				{
					json.Unmarshal,
					func(_ []byte, v interface{}) error {
						p := v.(*map[int]int)
						*p = make(map[int]int)
						(*p)[1] = 2
						(*p)[2] = 4
						return nil
					},
				},
			}
			patches := NewPatches()
			defer patches.Reset()
			for _, pair := range patchPairs {
				patches.ApplyFunc(pair[0], pair[1])
			}

			output, err := fake.Exec("", "")
			So(err, ShouldEqual, nil)
			So(output, ShouldEqual, outputExpect)

			var m map[int]int
			err = json.Unmarshal(nil, &m)
			So(err, ShouldEqual, nil)
			So(m[1], ShouldEqual, 2)
			So(m[2], ShouldEqual, 4)
		})

	})
}

func TestGetFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // 断言 DB.Get() 方法是否被调用

	m := NewMockDB(ctrl)
	m.EXPECT().Get(gomock.Eq("Tom")).Return(100, nil)

	assert.Equal(t, GetFromDB(m, "Tom"), 100)
}

// go convey
func TestMyModBaseConvey(t *testing.T) {
	Convey("TestMyModBaseConvey", t, func() {
		var (
			a    = 7
			b    = 7
			want = 0
		)

		So(MyMod(a, b), ShouldEqual, want)
	})

	Convey("TestMyModBaseConvey1", t, func() {
		var (
			a    = 7
			b    = 7
			want = 0
		)

		So(MyMod(a, b), ShouldEqual, want)
	})

	Convey("TestMyModBaseConvey2", t, func() {
		var (
			a    = 7
			b    = 7
			want = 0
		)

		So(MyMod(a, b), ShouldEqual, want)
	})

	Convey("TestMyModBaseConvey3", t, func() {
		Convey("TestMyModBaseConvey31", func() {
			var (
				a    = 7
				b    = 7
				want = 0
			)

			So(MyMod(a, b), ShouldEqual, want)
		})

		Convey("TestMyModBaseConvey32", func() {
			var (
				a    = 7
				b    = 7
				want = 0
			)

			So(MyMod(a, b), ShouldEqual, want)
		})
	})

}
