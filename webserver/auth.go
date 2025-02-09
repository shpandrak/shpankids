package webserver

import (
	"context"
	"net/http"
)

func doAuth(r *http.Request, w http.ResponseWriter) *string {
	session, _ := store.Get(r, "auth-session")

	// Check if the user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}

	email, ok := session.Values["email"].(string)
	if ok {
		r.Header.Set("X-User-Email", email)
		return &email
	}
	return nil

}

func doAuthOptional(r *http.Request) *string {
	session, _ := store.Get(r, "auth-session")

	// Check if the user is authenticated
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return nil
	}

	email, ok := session.Values["email"].(string)
	if ok {
		r.Header.Set("X-User-Email", email)
		return &email
	}
	return nil

}

func mustUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := doAuth(r, w)
		if s == nil || *s == "" {
			return
			//next.ServeHTTP(w, r)
		} else {
			// Add a value to the context
			ctx := context.WithValue(r.Context(), "x-shpankids-user", *s)
			// Pass the new context with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
