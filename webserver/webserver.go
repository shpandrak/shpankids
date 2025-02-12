package webserver

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	middleware "github.com/oapi-codegen/nethttp-middleware"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"shpankids/internal/api"
	"shpankids/openapi"
	"shpankids/shpankids"
	"shpankids/webserver/auth"
	"shpankids/webserver/auth/google"
)

const IndexPage = `
<html>
	<head>
		<title>Shpankids login</title>
	</head>
	<body>
		<h2>Login to system</h2>
		<p>
			Login with the following,
		</p>
		<ul>
			<li><a href="/login-gl">Google</a></li>
		</ul>
	</body>
</html>
`

var store = sessions.NewCookieStore([]byte("shpankids-secret"))

func Start(
	taskManager shpankids.TaskManager,
	userManager shpankids.UserManager,
	familyManager shpankids.FamilyManager,
	sessionManager shpankids.SessionManager,
) error {

	router := mux.NewRouter().StrictSlash(true)

	// Routes for the application
	router.HandleFunc("/login", handleMain)

	swagger, err := openapi.GetSwagger()
	if err != nil {
		return fmt.Errorf("error loading swagger spec\n: %w", err)
	}
	apiSubRouter := router.PathPrefix("/api").Subrouter()
	apiSubRouter.Use(middleware.OapiRequestValidator(swagger))
	//apiSubRouter.Use(mustUserMiddleware)

	apiImpl := api.NewOapiServerApiImpl(
		auth.GetUserInfo,
		userManager,
		taskManager,
		familyManager,
		sessionManager,
	)
	withStrictHandler := openapi.NewStrictHandlerWithOptions(apiImpl, nil, openapi.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: api.HandleErrors,
	})
	openapi.HandlerWithOptions(
		withStrictHandler,
		openapi.GorillaServerOptions{
			BaseURL:          "",
			BaseRouter:       router,
			Middlewares:      []openapi.MiddlewareFunc{mustUserMiddleware},
			ErrorHandlerFunc: api.HandleErrors,
		})

	// Register the Google OAuth routes
	google.RegisterCallbacks(router, func(r *http.Request, w http.ResponseWriter, email string) {
		// Create a session
		session, _ := store.Get(r, "auth-session")
		session.Values["authenticated"] = true
		session.Values["email"] = email
		err := session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		slog.Info(fmt.Sprintf("User logged in: %s", email))
		http.Redirect(w, r, "/ui", http.StatusTemporaryRedirect)
	})

	// demo UI and all its assets are served from /demo
	router.PathPrefix("/ui").Handler(http.StripPrefix("/ui", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		authEmail := doAuth(request, writer)
		if authEmail == nil {
			return
		}
		staticPath := "./ui/dist"

		// Clean the path to avoid directory traversal attacks as there is some
		// manual path manipulation below
		path := filepath.Clean(filepath.Join(staticPath, request.URL.Path))

		// Check if the path exists as a file
		_, err := os.Stat(path)

		// The path leads to a direct file, return a 404 if the file doesn't exist
		// otherwise as a SPA, we default to index.html
		if os.IsNotExist(err) {
			ext := filepath.Ext(path)

			if ext != "" {
				http.NotFound(writer, request)
				return
			}

			http.ServeFile(writer, request, filepath.Join(staticPath, "index.html"))
			return
		}

		if err != nil {
			// if we got an error (that wasn't that the file doesn't exist) stating the
			// file, return a 500 internal server error and stop
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// otherwise, use http.FileServer to serve the static file
		http.FileServer(http.Dir(staticPath)).ServeHTTP(writer, request)
	})))

	router.Path("/swagger").Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, "./ui/dist/swagger-ui.html")
	}))

	//router.Path("/metrics").Handler(promhttp.Handler())
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: router,
	}
	return srv.ListenAndServe()
}

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

/*
handleMain Function renders the index page when the application index route is called
*/
func handleMain(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(IndexPage))
}
