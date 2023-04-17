package restutilsgo

import (
	"fmt"
	"net/http"

	"github.com/jasonkofo/gocommon"
	"github.com/julienschmidt/httprouter"
)

const (
	SYSTEM_USER_ID     = 0
	SYSTEM_CLIENT_NAME = "peopleflow"
	SYSTEM_USERNAME    = "PEOPLEFLOW"
)

type IdentifierType int

const (
	IdentifierTypeEmail IdentifierType = 1 << iota
	IdentifierTypeIDPassport
	IdentifierTypeUsername
)

type HTTPRestServer interface {
	// GET allows you to register a given HTTP handler for a path. Your router
	// will therefore be able to register a "GET" endpoint to be handled by
	// "handler"
	GET(path string, handler UnauthenticatedHandler)
	// GETAuthenticated allows you to register a given HTTP handler for a path.
	// Your router will therefore be able to register an authenticated "GET"
	// endpoint to be handled by "handler"
	GETAuthenticated(path string, handler AuthenticatedHandler)
	// POST allows you to register a given HTTP handler for a path. Your router
	// will therefore be able to register a "POST" endpoint to be handled by
	// "handler"
	POST(path string, handler UnauthenticatedHandler)
	// POSTAuthenticated allows you to register a given HTTP handler for a path.
	// Your router will therefore be able to register an authenticated "POST"
	// endpoint to be handled by "handler"
	POSTAuthenticated(path string, handler AuthenticatedHandler)
	// PATCH allows you to register a given HTTP handler for a path. Your router
	// will therefore be able to register a "PATCH" endpoint to be handled by
	// "handler"
	PATCH(path string, handler UnauthenticatedHandler)
	// PATCHAuthenticated allows you to register a given HTTP handler for a path.
	// Your router will therefore be able to register an authenticated "PATCH"
	// endpoint to be handled by "handler"
	PATCHAuthenticated(path string, handler AuthenticatedHandler)
	// DELETE allows you to register a given HTTP handler for a path. Your router
	// will therefore be able to register a "DELETE" endpoint to be handled by
	// "handler"
	DELETE(path string, handler UnauthenticatedHandler)
	// DELETEAuthenticated allows you to register a given HTTP handler for a path.
	// Your router will therefore be able to register an authenticated "DELETE"
	// endpoint to be handled by "handler"
	DELETEAuthenticated(path string, handler AuthenticatedHandler)
	// PUT allows you to register a given HTTP handler for a path. Your router
	// will therefore be able to register a "PUT" endpoint to be handled by
	// "handler"
	PUT(path string, handler UnauthenticatedHandler)
	// PUTAuthenticated allows you to register a given HTTP handler for a path.
	// Your router will therefore be able to register an authenticated "PUT"
	// endpoint to be handled by "handler"
	PUTAuthenticated(path string, handler AuthenticatedHandler)
	Run() error
	// GetHandler returns the underlying native http.Handler method that is
	// found in your router
	GetHandler() http.Handler
	authenticatedFunction(handler AuthenticatedHandler) httprouter.Handle
	// wrapMiddlewares is a helper method that wraps the middlewares in a
	// protectionHandler
	wrapMiddlewares(handler UnauthenticatedHandler) httprouter.Handle
	// AddMiddlewares adds a function to be run as a post-authenticated
	// middleware
	AddMiddlewares(middleware ...MiddlewareFunc)
}

type httpRestServer struct {
	gocommon.LoggableBase
	port               int
	router             *httprouter.Router
	middlewares        []MiddlewareFunc
	authenticationFunc AuthFunc
}

type UserAccessCredentials struct {
	username string `json:"-"`
	userID   int    `json:"-"`
	Roles    []string
}

func (x *UserAccessCredentials) GetUsername() string {
	return x.username
}

func (x *UserAccessCredentials) GetUserID() int {
	return x.userID
}

func NewUserAccessCredentials(username string, userID int) *UserAccessCredentials {
	return &UserAccessCredentials{username: username, userID: userID}
}

func CreateSystemUserCredentials() *UserAccessCredentials {
	return NewUserAccessCredentials(SYSTEM_USERNAME, SYSTEM_USER_ID)
}

func (x *UserAccessCredentials) IsSystemUser() bool {
	return x.username == SYSTEM_USERNAME && x.userID == SYSTEM_USER_ID
}

type AuthFunc *func(r *http.Request) (*UserAccessCredentials, error)

// MiddlewareFunc is a helper type that assists the clients of these handlers in
// returning
type MiddlewareFunc func(r *http.Request) error

type AuthenticatedHandler func(context AuthenticatedHTTPServerContext)
type UnauthenticatedHandler func(context UnauthenticatedHTTPServerContext)

func CreateHTTPServer(port int, authFunc AuthFunc, logger gocommon.Logger) HTTPRestServer {
	return &httpRestServer{
		port:               port,
		router:             httprouter.New(),
		middlewares:        make([]MiddlewareFunc, 0),
		authenticationFunc: authFunc,
		LoggableBase:       gocommon.LoggableBase{Logger: logger},
	}
}

func (x *httpRestServer) GET(path string, handler UnauthenticatedHandler) {
	x.router.GET(path, x.wrapMiddlewares(handler))
}

func (x *httpRestServer) GETAuthenticated(path string, handler AuthenticatedHandler) {
	x.router.GET(path, x.authenticatedFunction(handler))
}

func (x *httpRestServer) POST(path string, handler UnauthenticatedHandler) {
	x.router.POST(path, x.wrapMiddlewares(handler))
}

func (x *httpRestServer) POSTAuthenticated(path string, handler AuthenticatedHandler) {
	x.router.POST(path, x.authenticatedFunction(handler))
}

func (x *httpRestServer) PATCH(path string, handler UnauthenticatedHandler) {
	x.router.PATCH(path, x.wrapMiddlewares(handler))
}

func (x *httpRestServer) PATCHAuthenticated(path string, handler AuthenticatedHandler) {
	x.router.PATCH(path, x.authenticatedFunction(handler))
}

func (x *httpRestServer) DELETE(path string, handler UnauthenticatedHandler) {
	x.router.DELETE(path, x.wrapMiddlewares(handler))
}

func (x *httpRestServer) DELETEAuthenticated(path string, handler AuthenticatedHandler) {
	x.router.DELETE(path, x.authenticatedFunction(handler))
}

func (x *httpRestServer) PUT(path string, handler UnauthenticatedHandler) {
	x.router.PUT(path, x.wrapMiddlewares(handler))
}

func (x *httpRestServer) PUTAuthenticated(path string, handler AuthenticatedHandler) {
	x.router.PUT(path, x.authenticatedFunction(handler))
}

func (x *httpRestServer) Run() error {
	if x.port == 0 {
		return fmt.Errorf("An invalid host was specified for this HTTP server")
	}
	x.Infof("Running local web server on port %d\n", x.port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", x.port), x.router); err != nil {
		return fmt.Errorf("Could not bind to port %v: %v", x.port, err)
	}

	return nil
}

// Use this if you want to use an alternate http server than the one referenced in the Run() method
func (x *httpRestServer) GetHandler() http.Handler {
	return x.router
}

func defaultAuthenticationFunction(r *http.Request) (*UserAccessCredentials, error) {
	// [JKG 2022-03-18] One day we will have a default authentication
	// function that should be entered into this branch
	return CreateSystemUserCredentials(), nil
}

// authenticatedFunction plays the role of the adapter between httprouter.Handle
// and a handler that contains the necessary abstractions and data that we will
// be able to use for our PeopleFlow authenticated handlers.
// Internally it calls the authentication methods that in turn call the
// metadata for the given requests
func (x *httpRestServer) authenticatedFunction(handler AuthenticatedHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		var (
			user *UserAccessCredentials
			err  error
			f    AuthFunc
		)

		if x.authenticationFunc != nil {
			f = x.authenticationFunc
		} else {
			fn := func(r *http.Request) (*UserAccessCredentials, error) { return defaultAuthenticationFunction(r) }
			f = &fn
		}
		if user, err = (*f)(r); err != nil {
			if httperror, ok := err.(*gocommon.HTTPError); ok {
				httperror.ToResultType().Send(w)
				return
			}
			gocommon.SendError(w, err)
			x.Error(err)
			return
		}

		ctx := newAuthenticatedServerContext(w, r, params, *user, x, x.Logger)
		defer ctx.ProtectionHandler(w)

		for i, middleware := range x.middlewares {
			if err := middleware(r); err != nil {
				gocommon.PanicServerErrorf("middleware level %v failed: %v", i, err)
			}
		}

		handler(ctx)
	}
}

func (x *httpRestServer) wrapMiddlewares(handler UnauthenticatedHandler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		ctx := newUnauthenticatedServerContext(w, r, params, x, x.Logger)
		defer ctx.ProtectionHandler(w)

		for i, middleware := range x.middlewares {
			if err := middleware(r); err != nil {
				gocommon.PanicServerErrorf("middleware level %v failed: %v", i, err)
			}
		}

		handler(ctx)
	}
}

func (x *httpRestServer) AddMiddlewares(middleware ...MiddlewareFunc) {
	x.middlewares = append(x.middlewares, middleware...)
}
