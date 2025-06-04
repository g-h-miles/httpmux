// Copyright 2024 Graham Miles. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

// Package httpmux provides a MultiRouter for routing requests to different
// routers based on path prefixes with automatic conflict detection.
//
// MultiRouter allows you to compose multiple routers together while maintaining
// clean separation of concerns. It automatically detects route conflicts at
// registration time to prevent shadowed routes.
//
// A simple example:
//
//	multi := httpmux.NewMultiRouter()
//
//	// API routes
//	apiRouter := httpmux.New()
//	apiRouter.GET("/users/{id}", UserHandler)
//	multi.Group("/api", apiRouter)
//
//	// Admin routes
//	adminRouter := multi.NewGroup("/admin")
//	adminRouter.GET("/dashboard", DashboardHandler)
//
//	// Frontend fallback
//	frontendRouter := httpmux.New()
//	frontendRouter.GET("/{path...}", SPAHandler)
//	multi.Default(frontendRouter)
//
//	http.ListenAndServe(":8080", multi)
//
// The MultiRouter will automatically detect conflicts like registering
// "/admin/users" in the default router when an "/admin" group already exists,
// and panic with a clear error message.

package httpmux

import (
	"fmt"
	"net/http"
	"strings"
)

// MultiRouter routes requests to different routers based on path prefixes
type MultiRouter struct {
	routes          map[string]*Router
	defaultRouter   *Router
	prefixes        []string // Keep track of prefixes in order for longest match
	registeredPaths []string // Track all paths registered in default router
	enableWarnings  bool
}

// NewMultiRouter creates a new MultiRouter
func NewMultiRouter() *MultiRouter {
	return &MultiRouter{
		routes:         make(map[string]*Router),
		prefixes:       make([]string, 0),
		enableWarnings: true,
	}
}

func (m *MultiRouter) Routes() map[string]*Router {
	return m.routes
}

// Group registers a router for a specific path prefix
func (m *MultiRouter) Group(prefix string, router *Router) {
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if prefix != "/" && strings.HasSuffix(prefix, "/") {
		prefix = prefix[:len(prefix)-1]
	}

	// Check conflicts - just call GetPaths() directly
	paths := router.getPaths()

	for _, path := range paths {
		fullPath := prefix + path

		// Check against all existing group prefixes
		for existingPrefix := range m.routes {
			if existingPrefix != "/" && existingPrefix != prefix && strings.HasPrefix(fullPath, existingPrefix) {
				panic(fmt.Sprintf("GROUP CONFLICT: Group '%s' route '%s' (full path: '%s') conflicts with existing group '%s'", prefix, path, fullPath, existingPrefix))
			}
		}
	}

	// Check existing groups
	for existingPrefix, existingRouter := range m.routes {
		if existingPrefix == "/" || existingPrefix == prefix {
			continue
		}

		existingPaths := existingRouter.getPaths()
		for _, existingPath := range existingPaths {
			fullExistingPath := existingPrefix + existingPath
			if strings.HasPrefix(fullExistingPath, prefix) {
				panic(fmt.Sprintf("GROUP CONFLICT: New group '%s' conflicts with existing route '%s' in group '%s'", prefix, fullExistingPath, existingPrefix))
			}
		}
	}

	m.routes[prefix] = router
	m.prefixes = append(m.prefixes, prefix)

	// Sort prefixes by length (longest first)
	for i := len(m.prefixes) - 1; i > 0; i-- {
		if len(m.prefixes[i]) > len(m.prefixes[i-1]) {
			m.prefixes[i], m.prefixes[i-1] = m.prefixes[i-1], m.prefixes[i]
		} else {
			break
		}
	}
}

// Default sets the default router for unmatched paths
func (m *MultiRouter) Default(router *Router) {
	// Get all paths from the router being set as default
	paths := router.getPaths()

	// Check each path against our group prefixes
	for _, path := range paths {
		for _, prefix := range m.prefixes {
			if prefix != "/" && strings.HasPrefix(path, prefix) {
				panic(fmt.Sprintf("ROUTE CONFLICT: Default router has route '%s' which conflicts with group '%s'! Move it to that group instead.", path, prefix))
			}
		}
	}

	m.defaultRouter = router
}

// ServeHTTP implements http.Handler
func (m *MultiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Find the longest matching prefix
	for _, prefix := range m.prefixes {
		if prefix == "/" {
			continue
		}

		if strings.HasPrefix(path, prefix) {
			router := m.routes[prefix]

			// Strip prefix from path
			originalPath := r.URL.Path
			newPath := strings.TrimPrefix(path, prefix)
			if newPath == "" {
				newPath = "/"
			}
			r.URL.Path = newPath

			router.ServeHTTP(w, r)

			// Restore original path
			r.URL.Path = originalPath
			return
		}
	}

	// Check for root prefix "/"
	if rootRouter := m.routes["/"]; rootRouter != nil {
		rootRouter.ServeHTTP(w, r)
		return
	}

	// Before using default router, check if path conflicts with any group prefix
	if m.defaultRouter != nil {
		for _, prefix := range m.prefixes {
			if prefix != "/" && strings.HasPrefix(path, prefix) {
				panic(fmt.Sprintf("ROUTE CONFLICT: Path '%s' should be in group '%s', not default router!", path, prefix))
			}
		}

		m.defaultRouter.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

// Convenience method to create a new router for a group
func (m *MultiRouter) NewGroup(prefix string) *Router {
	router := New()
	m.Group(prefix, router)
	return router
}

// Add method to register routes in default router with conflict checking
func (m *MultiRouter) RegisterDefault(method, path string, handler http.HandlerFunc) {
	// Check if path conflicts with any existing group prefix
	for _, prefix := range m.prefixes {
		if prefix != "/" && strings.HasPrefix(path, prefix) {
			panic(fmt.Sprintf("ROUTE CONFLICT: Cannot register '%s' - conflicts with group '%s'", path, prefix))
		}
	}

	if m.defaultRouter == nil {
		m.defaultRouter = New()
	}

	m.registeredPaths = append(m.registeredPaths, path)
	m.defaultRouter.HandleFunc(method, path, handler)
}
