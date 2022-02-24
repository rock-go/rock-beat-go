//go:build windows
// +build windows

/*
Package wmi provides a WQL interface for WMI on Windows.

Example code to print names of running processes:

	type Win32_Process struct {
		Name string
	}

	func main() {
		var dst []Win32_Process
		q := wmi.CreateQuery(&dst, "")
		err := wmi.Query(q, &dst)
		if err != nil {
			log.Fatal(err)
		}
		for i, v := range dst {
			println(i, v.Name)
		}
	}

*/
package wmi

import (
	"errors"
	"fmt"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"sync"
)

var (
	ErrInvalidEntityType = errors.New("wmi: invalid entity type")
	// ErrNilCreateObject is the error returned if CreateObject returns nil even
	// if the error was nil.
	ErrNilCreateObject = errors.New("wmi: create object returned nil")
	lock               sync.Mutex
)

// S_FALSE is returned by CoInitializeEx if it was already called on this thread.
const S_FALSE = 0x00000001

// Query runs the WQL query and appends the values to dst.
//
// dst must have type *[]S or *[]*S, for some struct type S. Fields selected in
// the query must have the same name in dst. Supported types are all signed and
// unsigned integers, time.Time, string, bool, or a pointer to one of those.
// Array types are not supported.
//
// By default, the local machine and default namespace are used. These can be
// changed using connectServerArgs. See
// https://docs.microsoft.com/en-us/windows/desktop/WmiSdk/swbemlocator-connectserver
// for details.
//

// CallMethod calls a method named methodName on an instance of the class named
// className, with the given params.
//
// CallMethod is a wrapper around DefaultClient.CallMethod.
func CallMethod(connectServerArgs []interface{}, className, methodName string, params []interface{}) (int32, error) {
	return DefaultClient.CallMethod(connectServerArgs, className, methodName, params)
}

// A Client is an WMI query client.
//
// Its zero value (DefaultClient) is a usable client.
type Client struct {
	// NonePtrZero specifies if nil values for fields which aren't pointers
	// should be returned as the field types zero value.
	//
	// Setting this to true allows stucts without pointer fields to be used
	// without the risk failure should a nil value returned from WMI.
	NonePtrZero bool

	// PtrNil specifies if nil values for pointer fields should be returned
	// as nil.
	//
	// Setting this to true will set pointer fields to nil where WMI
	// returned nil, otherwise the types zero value will be returned.
	PtrNil bool

	// AllowMissingFields specifies that struct fields not present in the
	// query result should not result in an error.
	//
	// Setting this to true allows custom queries to be used with full
	// struct definitions instead of having to define multiple structs.
	AllowMissingFields bool

	// SWbemServiceClient is an optional SWbemServices object that can be
	// initialized and then reused across multiple queries. If it is null
	// then the method will initialize a new temporary client each time.
}

// DefaultClient is the default Client and is used by Query, QueryNamespace, and CallMethod.
var DefaultClient = &Client{}

// coinitService coinitializes WMI service. If no error is returned, a cleanup function
// is returned which must be executed (usually deferred) to clean up allocated resources.
func (c *Client) coinitService(connectServerArgs ...interface{}) (*ole.IDispatch, func(), error) {
	var unknown *ole.IUnknown
	var wmi *ole.IDispatch
	var serviceRaw *ole.VARIANT

	// be sure teardown happens in the reverse
	// order from that which they were created
	deferFn := func() {
		if serviceRaw != nil {
			serviceRaw.Clear()
		}
		if wmi != nil {
			wmi.Release()
		}
		if unknown != nil {
			unknown.Release()
		}
		ole.CoUninitialize()
	}

	// if we error'ed here, clean up immediately
	var err error
	defer func() {
		if err != nil {
			deferFn()
		}
	}()

	err = ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED)
	if err != nil {
		oleCode := err.(*ole.OleError).Code()
		if oleCode != ole.S_OK && oleCode != S_FALSE {
			return nil, nil, err
		}
	}

	unknown, err = oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		return nil, nil, err
	} else if unknown == nil {
		return nil, nil, ErrNilCreateObject
	}

	wmi, err = unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, nil, err
	}

	// service is a SWbemServices
	serviceRaw, err = oleutil.CallMethod(wmi, "ConnectServer", connectServerArgs...)
	if err != nil {
		return nil, nil, err
	}

	return serviceRaw.ToIDispatch(), deferFn, nil
}

// CallMethod calls a WMI method named methodName on an instance
// of the class named className. It passes in the arguments given
// in params. Use connectServerArgs to customize the machine and
// namespace; by default, the local machine and default namespace
// are used. See
// https://docs.microsoft.com/en-us/windows/desktop/WmiSdk/swbemlocator-connectserver
// for details.
func (c *Client) CallMethod(connectServerArgs []interface{}, className, methodName string, params []interface{}) (int32, error) {
	service, cleanup, err := c.coinitService(connectServerArgs...)
	if err != nil {
		return 0, fmt.Errorf("coinit: %v", err)
	}
	defer cleanup()

	// Get class
	classRaw, err := oleutil.CallMethod(service, "Get", className)
	if err != nil {
		return 0, fmt.Errorf("CallMethod Get class %s: %v", className, err)
	}
	class := classRaw.ToIDispatch()
	defer classRaw.Clear()

	// Run method
	resultRaw, err := oleutil.CallMethod(class, methodName, params...)
	if err != nil {
		return 0, fmt.Errorf("CallMethod %s.%s: %v", className, methodName, err)
	}
	resultInt, ok := resultRaw.Value().(int32)
	if !ok {
		return 0, fmt.Errorf("return value was not an int32: %v (%T)", resultRaw, resultRaw)
	}

	return resultInt, nil
}

func oleInt64(item *ole.IDispatch, prop string) (int64, error) {
	v, err := oleutil.GetProperty(item, prop)
	if err != nil {
		return 0, err
	}
	defer v.Clear()

	i := int64(v.Val)
	return i, nil
}
