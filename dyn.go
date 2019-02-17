// Package dyn provides late binding helpers for go. It uses
// reflection to implement these functions and provides interfaces that a
// user may use to override the default go behaviors. This allows for one
// to build late bound language extensions for go.
package dyn

import (
	"errors"
	"fmt"
	"reflect"
)

// Applier is any type that knows how to apply arguments to its self
// and return a value.
type Applier interface {
	Apply(args ...interface{}) interface{}
}

// Apply will apply the type with the supplied arguments. If the type
// is an Applier then Apply will be called. Otherwise if the type is a go
// function type then it will be called with reflect in the standard
// way. Any other type will panic.
func Apply(f interface{}, args ...interface{}) interface{} {
	if a, ok := f.(Applier); ok {
		return a.Apply(args...)
	}
	return apply(reflect.ValueOf(f), args...)
}

func apply(fnv reflect.Value, args ...interface{}) interface{} {
	fnt := fnv.Type()
	argvs := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg == nil {
			fnint := fnt.In(i)
			fnink := fnint.Kind()
			switch fnink {
			case reflect.Chan, reflect.Func,
				reflect.Interface, reflect.Map,
				reflect.Ptr, reflect.Slice:
				argvs[i] = reflect.Zero(fnint)
			default:
				// this will cause a panic but that is what is
				// intended
				argvs[i] = reflect.ValueOf(arg)
			}
		} else {
			argvs[i] = reflect.ValueOf(arg)
		}
	}
	outvs := fnv.Call(argvs)
	switch len(outvs) {
	case 0:
		return nil
	case 1:
		return outvs[0].Interface()
	default:
		outs := make([]interface{}, len(outvs))
		for i, outv := range outvs {
			outs[i] = outv.Interface()
		}
		return outs
	}
}

// Bind will create a context in which the function application is
// deferrred. When the returned context is called the function is applied
// and the result returned.
func Bind(fn interface{}, args ...interface{}) func() interface{} {
	return func() interface{} {
		return Apply(fn, args...)
	}
}

// Finder is any type that can index its self and return a value it
// contains and whether a value was at that index.
type Finder interface {
	Find(interface{}) (interface{}, bool)
}

// Find looks up a value in an associative object. If the type of the
// assocObj is a Finder then Find will be called and the value
// returned. Otherwise reflection will be used to do a lookup on native
// go types. If the type is a struct it may be indexed by an integer or a
// string, any other index type will panic. If the type is a map then the
// selector will be treated as a key, if the key is of the wrong type
// then Find will panic. If the type is a slice then the selector must be
// an int, if the index is in the slice then a value is returned,
// otherwise nil and false will be returned. If the type is a pointer to
// any of the above then the pointer will be dereferenced and then the
// above semantics will hold.
func Find(assocObj interface{}, selector interface{}) (interface{}, bool) {
	o, ok := assocObj.(Finder)
	if ok {
		return o.Find(selector)
	}
	return findReflect(reflect.ValueOf(assocObj), selector)
}

func findReflect(objv reflect.Value, selector interface{}) (interface{}, bool) {
	switch objv.Kind() {
	case reflect.Struct:
		switch s := selector.(type) {
		case int:
			if s < 0 || s >= objv.NumField() {
				return nil, false
			}
			return objv.Field(s).Interface(), true
		case string:
			out := objv.FieldByName(s)
			if !out.IsValid() {
				return nil, false
			}
			return out.Interface(), true
		default:
			panic(errors.New("structs can only be referenced by index or name"))
		}
	case reflect.Map:
		out := objv.MapIndex(reflect.ValueOf(selector))
		if !out.IsValid() {
			return nil, false
		}
		return out.Interface(), true
	case reflect.Slice:
		idx := selector.(int)
		if idx < 0 || idx >= objv.Len() {
			return nil, false
		}
		return objv.Index(idx).Interface(), true
	case reflect.Ptr:
		return findReflect(objv.Elem(), selector)
	default:
		panic(errors.New("Find passed a non associative type"))
	}
}

// At uses Find to retieve an object but ignores whether the value was
// found.
func At(assocObj interface{}, selector interface{}) interface{} {
	out, _ := Find(assocObj, selector)
	return out
}

// MessageReceivers are any object that implements its own messaging
// semantics.
type MessageReceiver interface {
	Receive(message ...interface{}) interface{}
}

// ErrDoesNotUnderstand is returned when a method can not be located
// for a message.
type ErrDoesNotUnderstand struct {
	o       interface{}
	message []interface{}
}

func (e ErrDoesNotUnderstand) Error() string {
	return fmt.Sprintf("Object %v does not understand %v", e.o, e.message)
}

// DoesNotUnderstand is a constructor for an ErrDoesNotUnderstand error.
func DoesNotUnderstand(o interface{}, message ...interface{}) error {
	return ErrDoesNotUnderstand{
		o:       o,
		message: message,
	}
}

// Send provides a late binding way to call methods on objects. It
// abstracts the method call semantics and allows user defined types to
// implement their own message semantics. Send will send a message to a
// receiver and return the returned value. If the receiver is a
// MessageReceiver then Receive will be called. For any other type the go
// method will be looked up by name based on the first element of message
// and the method will be applied with the rest of the message.
func Send(rcvr interface{}, message ...interface{}) interface{} {
	r, ok := rcvr.(MessageReceiver)
	if ok {
		return r.Receive(message...)
	}
	rcvrv := reflect.ValueOf(rcvr)
	method := rcvrv.MethodByName(message[0].(string))
	if !method.IsValid() {
		panic(DoesNotUnderstand(rcvr, message...))
	}
	return apply(method, message[1:]...)
}