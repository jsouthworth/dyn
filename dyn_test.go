package dyn

import (
	"errors"
	"fmt"
	"testing"
)

func ExampleApply() {
	fmt.Println(Apply(func(x int) int { return x * x }, 10))
	// Output: 100
}

func ExampleApply_map() {
	fmt.Println(Apply(map[string]int{"a": 10, "b": 20, "c": 30}, "a"))
	// Output: 10
}

func ExampleApply_slice() {
	fmt.Println(Apply([]string{"foo", "bar", "baz"}, 1))
	// Output: bar
}

func ExampleApply_struct() {
	type example struct {
		Foo, Bar, Baz string
	}
	fmt.Println(Apply(example{"foo", "bar", "baz"}, "Foo"))
	fmt.Println(Apply(example{"foo", "bar", "baz"}, 1))
	// Output: foo
	// bar
}

type annotate struct {
	name string
	fn   interface{}
}

func (a annotate) Apply(args ...interface{}) interface{} {
	return fmt.Sprintf("%s(%v)->%v", a.name, args, Apply(a.fn, args...))
}

func ExampleApply_applier() {
	fmt.Println(Apply(annotate{name: "square", fn: func(x int) int {
		return x * x
	}}, 10))
	// Output: square([10])->100
}

func ExampleApply_multipleArgs() {
	fmt.Println(Apply(func(x int, y string) string {
		return fmt.Sprintf("%s:%d", y, x)
	}, 10, "foo"))
	// Output: foo:10
}

func ExampleApply_multipleReturns() {
	out := Apply(func(x int) (int, int) { return x, x + 1 }, 10)
	fmt.Println(At(out, 0), At(out, 1))
	// Output: 10 11
}

func ExampleApply_noReturn() {
	fmt.Println(Apply(func(x int) { fmt.Println(x) }, 10))
	// Output: 10
	//<nil>
}

func ExampleApply_nilArg() {
	fmt.Println(Apply(func(err error) error {
		if err == nil {
			return errors.New("oops!")
		}
		return err
	}, nil))

	// Output: oops!
}

type receiver struct {
}

func (r *receiver) String() string {
	return "rcvr!"
}

func ExampleSend() {
	fmt.Println(Send(&receiver{}, "String"))
	// Output: rcvr!
}

type class struct {
	super        *class
	methods      map[string]interface{}
	instanceVars []string
}

func newClass(super *class, methods map[string]interface{}, instanceVars []string) *class {
	return &class{
		super:        super,
		methods:      methods,
		instanceVars: instanceVars,
	}
}
func (c *class) lookupMethod(selector string) (interface{}, bool) {
	if c == nil {
		return nil, false
	}
	method, ok := c.methods[selector]
	if ok {
		return method, ok
	}
	return c.super.lookupMethod(selector)
}

func (c *class) parseMessage(message ...interface{}) (interface{}, []interface{}, bool) {

	method, ok := c.lookupMethod(message[0].(string))
	if !ok || method == nil {
		return nil, nil, false
	}
	return method, message[1:], true
}

func (c *class) matchInstanceVar(name string, data []interface{}) (interface{}, bool, int) {
	if c == nil {
		return nil, false, 0
	}
	out, ok, idx := c.super.matchInstanceVar(name, data)
	if ok {
		return out, ok, idx
	}
	for _, ivName := range c.instanceVars {
		if ivName == name {
			return data[idx], true, idx
		}
		idx++
	}
	return nil, false, idx
}

func (c *class) lenInstanceVars() int {
	if c == nil {
		return 0
	}
	return len(c.instanceVars) + c.super.lenInstanceVars()
}

func (c *class) New(data ...interface{}) *object {
	ivars := make([]interface{}, c.lenInstanceVars())
	copy(ivars, data)
	return &object{class: c, data: ivars}
}

type object struct {
	class *class
	super *object
	data  []interface{}
}

func (o *object) Receive(message ...interface{}) interface{} {
	method, args, understood := o.class.parseMessage(message...)
	if !understood {
		panic(DoesNotUnderstand(o, message...))
	}
	return Apply(method, PrependArg(o, args...)...)
}

func (o *object) Find(name interface{}) (interface{}, bool) {
	if name == "super" {
		return o.super, true
	}
	out, found, _ := o.class.matchInstanceVar(name.(string), o.data)
	return out, found
}

func ExampleSend_inheritance() {
	// OK, this is a little perverse but it shows the power
	// of this stuff... Even this is more fixed than it needs
	// to be but it is just an example.
	fooCl := newClass(nil, map[string]interface{}{
		"string": func(self *object) interface{} {
			return At(self, "a")
		},
		"other": func(self *object) interface{} {
			return Send(self, "string")
		},
	}, []string{"a"})
	barCl := newClass(fooCl, map[string]interface{}{
		"string": func(self *object) interface{} {
			return At(self, "b")
		},
	}, []string{"b"})
	foo := fooCl.New("foo")
	fmt.Println(Send(foo, "other"))
	bar := barCl.New("bar", "quux")
	fmt.Println(Send(bar, "other"))

	// Output: foo
	// quux
}

func ExampleAt_struct() {
	type test struct {
		Foo string
	}

	fmt.Println(At(&test{Foo: "hello1"}, "Foo"))
	fmt.Println(At(test{Foo: "hello2"}, "Foo"))
	fmt.Println(At(test{Foo: "hello3"}, 0))
	// Output: hello1
	// hello2
	// hello3
}

func ExampleAt_map() {
	fmt.Println(At(map[string]string{"foo": "hello"}, "foo"))
	// Output: hello
}

func ExampleAt_slice() {
	fmt.Println(At([]string{"hello"}, 0))
	// Output: hello
}

func ExampleEqual_goNative() {
	fmt.Println(Equal(1, 2))
	// Output: false
}

type equalExample struct {
	m map[string]int
}

func (e *equalExample) Equal(other interface{}) bool {
	oex, isEqualExample := other.(*equalExample)
	if !isEqualExample {
		return false
	}
	for k, v := range e.m {
		if oex.m[k] != v {
			return false
		}
	}
	return true
}

func ExampleEqual_equaler() {
	a := &equalExample{m: map[string]int{"a": 10, "b": 20, "c": 30}}
	b := &equalExample{m: map[string]int{"a": 10, "b": 20, "c": 30}}
	fmt.Println(Equal(a, b))
	//Output: true
}

func ExampleEqual_oneEqualer() {
	a := &equalExample{m: map[string]int{"a": 10, "b": 20, "c": 30}}
	b := map[string]int{"a": 10, "b": 20, "c": 30}
	fmt.Println(Equal(b, a))
	//Output: false
}

func ExampleCompare_goNative() {
	fmt.Println(Compare("a", "b"))
	//Output: -1
}

type invertedCompare int

func (e invertedCompare) Compare(other interface{}) int {
	// Note: this version of compare inverts the standard relationship
	oe := other.(invertedCompare)
	switch {
	case e > oe:
		return -1
	case e < oe:
		return 1
	default:
		return 0
	}

}
func ExampleCompare_comparer() {
	fmt.Println(Compare(invertedCompare(1), invertedCompare(2)))
	//Output: 1
}

func ExampleCompose() {
	square := func(x int) int {
		return x * x
	}
	negate := func(x int) int {
		return -x
	}
	fmt.Println(Apply(Compose(negate, square), 5))
	//Output: -25
}

func ExampleCompse_multiReturn() {
	double := func(x interface{}) (interface{}, interface{}) {
		return x, x
	}
	takeTwo := func(x, y int) bool {
		return x == y
	}
	fmt.Println(Apply(Compose(double, takeTwo, double), 5))
	// Output: [true true]
}

var argList = []interface{}{1, 2, 3, 4, 5, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40}

func BenchmarkPrependArg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = PrependArg("r", argList...)
	}
}

func BenchmarkPrependWithAppend(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = append([]interface{}{"r"}, argList...)
	}
}

func ExamplePrependArg() {
	args := []interface{}{"b", "c", "d", "e"}
	args = PrependArg("a", args...)
	fmt.Println(args)
	// Output: [a b c d e]
}
