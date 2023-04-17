package restutilsgo

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/jasonkofo/gocommon"
	"github.com/julienschmidt/httprouter"
	"github.com/martinlindhe/base36"
	"google.golang.org/protobuf/proto"
)

type UnauthenticatedHTTPServerContext interface {
	ReadJSON(obj interface{})
	ReadProtoJSON(obj proto.Message)
	SendJSON(obj interface{})
	SendBytes(bytes []byte)
	SendID(id interface{})
	SendOK()
	SendPong()
	ReadRouterParamByName(name string) string
	RequestContext() context.Context
	ReadQueryParameterAsString(keys ...string) string
	ProtectionHandler(w http.ResponseWriter)
}

type AuthenticatedHTTPServerContext interface {
	UnauthenticatedHTTPServerContext
	UserCredentials() UserAccessCredentials
	GetUserID() int
	GetUsername() string
	IsSystemUser() bool
}

func newUnauthenticatedServerContext(
	w http.ResponseWriter,
	r *http.Request,
	params httprouter.Params,
	server HTTPRestServer,
	logger gocommon.Logger,
) UnauthenticatedHTTPServerContext {
	now := time.Now()
	prefix := base36.Encode(uint64(now.Unix()))
	return &httpServerContext{
		response:     w,
		request:      r,
		server:       server,
		prefix:       prefix,
		routerParams: &params,
		timestamp:    &now,
		LoggableBase: gocommon.LoggableBase{Logger: logger},
	}
}

func (x *httpServerContext) ReadJSON(obj interface{}) {
	x.Debugf("[%v] Attempting to read JSON from request body", x.prefix)
	readJSON(x.request, obj)
	if b, err := json.Marshal(obj); err == nil {
		x.Debugf("request body: %v", string(b))
	}
}

func (x *httpServerContext) ReadProtoJSON(obj proto.Message) { readProtoJSON(x.request, obj) }

func (x *httpServerContext) SendJSON(obj interface{}) {
	jsonbytes := sendJSON(x.response, obj)
	x.Infof("[%v] Successfully sent JSON response: %v", x.prefix, string(jsonbytes))
}

func (x *httpServerContext) SendBytes(bytes []byte) {
	sendBytes(x.response, bytes)
	x.Infof("[%v] Successfully sent JSON response: %v", x.prefix, string(bytes))
}

func (x *httpServerContext) SendID(id interface{}) {
	jsonbytes := sendID(x.response, id)
	x.Infof("[%v] Successfully sent ID response: %v", x.prefix, string(jsonbytes))
}
func (x *httpServerContext) SendOK() {
	jsonbytes := sendOK(x.response)
	x.Infof("[%v] Successfully sent OK response: %v", x.prefix, string(jsonbytes))
}

func (x *httpServerContext) SendPong() {
	jsonbytes := sendPong(x.response)
	x.Infof("[%v] Successfully sent Pong response: %v", x.prefix, string(jsonbytes))
}

func (x *httpServerContext) RequestContext() context.Context { return x.request.Context() }
func (x *httpServerContext) ReadRouterParamByName(name string) string {
	return x.routerParams.ByName(name)
}
func (x *httpServerContext) ReadQueryParameterAsString(keys ...string) string {
	return readQueryParameterAsString(x.request, keys...)
}

func (x *httpServerContext) ProtectionHandler(w http.ResponseWriter) {
	if e := recover(); e != nil {
		var res *gocommon.ResultType[interface{}]
		switch err := e.(type) {
		case string:
			res = gocommon.SendErrorString(w, err)
		case gocommon.HTTPError:
			res = err.ToResultType()
			res.Send(w)
		case *gocommon.HTTPError:
			res = err.ToResultType()
			res.Send(w)
		case error:
			gocommon.SendError(w, err)
		}
		res.Stack = string(debug.Stack())
		x.Errorf("[%v] %v", x.prefix, res.String)
	}
}

func newAuthenticatedServerContext(
	w http.ResponseWriter,
	r *http.Request,
	params httprouter.Params,
	user UserAccessCredentials,
	server HTTPRestServer,
	logger gocommon.Logger,
) AuthenticatedHTTPServerContext {
	unauthedctx := newUnauthenticatedServerContext(w, r, params, server, logger)
	return &authenticatedHTTPServerContext{
		UnauthenticatedHTTPServerContext: unauthedctx,
		userCredentials:                  user,
		LoggableBase:                     gocommon.LoggableBase{Logger: logger},
	}
}

// HTTPServerContext is passed into PeopleFlow request handlers
type httpServerContext struct {
	gocommon.LoggableBase
	prefix       string
	server       HTTPRestServer
	response     http.ResponseWriter
	request      *http.Request
	routerParams *httprouter.Params
	timestamp    *time.Time
}

type authenticatedHTTPServerContext struct {
	gocommon.LoggableBase
	UnauthenticatedHTTPServerContext
	userCredentials UserAccessCredentials
}

func (x *authenticatedHTTPServerContext) UserCredentials() UserAccessCredentials {
	y := x.userCredentials
	return y
}

func (x *authenticatedHTTPServerContext) GetUserID() int {
	y := x.UserCredentials()
	return y.GetUserID()
}

func (x *authenticatedHTTPServerContext) GetUsername() string {
	y := x.UserCredentials()
	return y.GetUsername()
}

func (x *authenticatedHTTPServerContext) IsSystemUser() bool {
	y := x.UserCredentials()
	return y.IsSystemUser()
}
