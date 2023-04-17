package restutilsgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/jasonkofo/gocommon"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Validator interface {
	Validate() error
}

// ParseID parses a 64-bit integer, and returns zero on failure.
func ParseID(s string) int64 {
	id, _ := strconv.ParseInt(s, 10, 64)
	return id
}

// readJSON reads the body of the request, and unmarshals it into 'obj'.
// if the object that was decoded implements the Validator interface, then
// the validation of the object will be handled in this method
func readJSON(r *http.Request, obj interface{}) {
	if r.Body == nil {
		gocommon.PanicBadRequestf("(readJSON) Request body is empty")
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(obj)
	if err != nil {
		gocommon.PanicBadRequestf("(readJSON) Failed to decode JSON - %v", err.Error())
	}
	if obj, ok := obj.(Validator); ok {
		if err := obj.Validate(); err != nil {
			gocommon.PanicBadRequestf("ReadJSON failed: Failed to validate dto data - %v", err.Error())
		}
	}
}

func readProtoJSON(r *http.Request, obj proto.Message) {
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		gocommon.Panic(err)
	}
	if err := protojson.Unmarshal(bodyBytes, obj); err != nil {
		gocommon.Panic(err)
	}
}

// readQueryParameterAsString is a small helper method that allows you to
// read the string value from the request object. In the case where there is
// more than one value stored at the key, then only the first entry will be
// returned
func readQueryParameterAsString(r *http.Request, keys ...string) string {
	queryvalues := r.URL.Query()
	for _, key := range keys {
		value := queryvalues.Get(key)
		if value != "" {
			return value
		}
	}
	return ""
}

func sendJSON[TValue any](w http.ResponseWriter, obj TValue) []byte {
	w.Header().Set("Content-Type", "application/json")
	x := gocommon.ResultType[TValue]{
		StatusCode: http.StatusOK,
		Data:       &obj,
		Result:     gocommon.SuccessValidResultType,
	}
	rv := reflect.ValueOf(obj)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Array || rv.Kind() == reflect.Slice {
		if _, ok := rv.Interface().([]byte); !ok {
			count := rv.Len()
			x.Count = &count
		}
	}
	b, _ := json.Marshal(x)
	w.Write(b)
	return b
}

// SendText converts text (of type string) into a the array and sends it as an
// HTTP text/plain response
func sendText(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "application/json")
	x := gocommon.ResultType[interface{}]{
		StatusCode: http.StatusOK,
		Message:    text,
		Result:     gocommon.SuccessValidResultType,
	}
	b, _ := json.Marshal(x)
	w.Write(b)
}

// SendID encodes 'id' as a string, and sends it as an HTTP text/plain response.
func sendID(w http.ResponseWriter, id interface{}) []byte {
	w.Header().Set("Content-Type", "application/json")
	x := gocommon.ResultType[map[string]interface{}]{
		StatusCode: http.StatusOK,
		Data:       &map[string]interface{}{"id": id},
		Result:     gocommon.SuccessValidResultType,
	}
	b, _ := json.Marshal(x)
	w.Write(b)
	return b
}

// SendOK sends "OK" as a text/plain response.
func sendOK(w http.ResponseWriter) []byte {
	w.Header().Set("Content-Type", "application/json")
	x := gocommon.ResultType[interface{}]{
		StatusCode: http.StatusOK,
		Result:     gocommon.SuccessValidResultType,
	}
	b, _ := json.Marshal(x)
	w.Write(b)
	return b
}

// SendBytes sends text as an HTTP text/plain response
func sendBytes(w http.ResponseWriter, bytes []byte) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write(bytes)
}

// SendPong sends a reply to an HTTP ping request, which checks if the service
// is alive.
func sendPong(w http.ResponseWriter) []byte {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=0, no-cache")
	b := []byte(fmt.Sprintf(`{"Timestamp": %v}`, time.Now().Unix()))
	w.Write(b)
	return b
}

func GetJSON(url string, obj interface{}) error {
	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return fmt.Errorf("a pointer reference must be passed into GetJSON")
	}
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to perform GET HTTP Request: %v", err)
	}
	defer resp.Body.Close()
	if err != nil {
		return fmt.Errorf("could not read body of response data: %v", err)
	}
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(obj); err != nil {
		gocommon.PanicBadRequestf("GetJSON failed: Failed to decode JSON - %v", err.Error())
	}

	return nil
}
