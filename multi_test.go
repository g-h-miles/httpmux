package httpmux

import (
	"strings"
	"testing"
)

func TestMultiRouter_NoConflicts(t *testing.T) {
	multi := NewMultiRouter()

	// Create API router
	apiRouter := New()
	apiRouter.GET("/users/{id}", dummyHandler)
	apiRouter.GET("/files/{filepath...}", dummyHandler)
	multi.Group("/api", apiRouter)

	// Create admin router
	adminRouter := multi.NewGroup("/admin")
	adminRouter.GET("/dashboard", dummyHandler)
	adminRouter.GET("/users", dummyHandler)

	// Create default router
	defaultRouter := New()
	defaultRouter.GET("/home", dummyHandler)
	defaultRouter.GET("/about", dummyHandler)
	multi.Default(defaultRouter)

	// Should not panic
}

func TestMultiRouter_DefaultRouterConflictsWithGroup(t *testing.T) {
	multi := NewMultiRouter()

	// Create admin group first
	adminRouter := multi.NewGroup("/admin")
	adminRouter.GET("/users", dummyHandler)

	// Try to add conflicting route to default router
	defaultRouter := New()
	defaultRouter.GET("/admin/dashboard", dummyHandler) // This should conflict

	defer func() {
		if r := recover(); r != nil {
			msg := r.(string)
			if !strings.Contains(msg, "ROUTE CONFLICT") || !strings.Contains(msg, "/admin/dashboard") {
				t.Errorf("Expected conflict error for /admin/dashboard, got: %v", r)
			}
		} else {
			t.Error("Expected panic for conflicting default route")
		}
	}()

	multi.Default(defaultRouter) // Should panic
}

func TestMultiRouter_GroupConflictsWithExistingGroup(t *testing.T) {
	multi := NewMultiRouter()

	// Create API router with admin route
	apiRouter := New()
	apiRouter.GET("/admin/users", dummyHandler) // This creates /api/admin/users
	multi.Group("/api", apiRouter)

	// Try to create conflicting group
	adminRouter := New()
	adminRouter.GET("/users", dummyHandler) // This would create /api/admin/users

	defer func() {
		if r := recover(); r != nil {
			msg := r.(string)
			if !strings.Contains(msg, "GROUP CONFLICT") {
				t.Errorf("Expected group conflict error, got: %v", r)
			}
		} else {
			t.Error("Expected panic for conflicting group")
		}
	}()

	multi.Group("/api/admin", adminRouter) // Should panic
}

func TestMultiRouter_NewGroupConflictsWithExistingRoute(t *testing.T) {
	multi := NewMultiRouter()

	// Create API router with specific route
	apiRouter := New()
	apiRouter.GET("/v2/users", dummyHandler) // This creates /api/v2/users
	multi.Group("/api", apiRouter)

	// Try to create group that would shadow existing route
	v2Router := New()
	v2Router.GET("/users", dummyHandler) // This would also create /api/v2/users

	defer func() {
		if r := recover(); r != nil {
			msg := r.(string)
			if !strings.Contains(msg, "GROUP CONFLICT") {
				t.Errorf("Expected group conflict error, got: %v", r)
			}
		} else {
			t.Error("Expected panic for conflicting new group")
		}
	}()

	multi.Group("/api/v2", v2Router) // Should panic
}

func TestMultiRouter_MultipleDefaultConflicts(t *testing.T) {
	multi := NewMultiRouter()

	// Create multiple groups
	apiRouter := multi.NewGroup("/api")
	apiRouter.GET("/users", dummyHandler)

	adminRouter := multi.NewGroup("/admin")
	adminRouter.GET("/dashboard", dummyHandler)

	// Default router with multiple conflicts
	defaultRouter := New()
	defaultRouter.GET("/api/files", dummyHandler)   // Conflicts with /api
	defaultRouter.GET("/admin/users", dummyHandler) // Conflicts with /admin
	defaultRouter.GET("/home", dummyHandler)        // No conflict

	defer func() {
		if r := recover(); r != nil {
			msg := r.(string)
			// Should catch the first conflict (/api/files)
			if !strings.Contains(msg, "ROUTE CONFLICT") || !strings.Contains(msg, "/api/files") {
				t.Errorf("Expected conflict error for /api/files, got: %v", r)
			}
		} else {
			t.Error("Expected panic for conflicting default routes")
		}
	}()

	multi.Default(defaultRouter) // Should panic on first conflict
}

func TestMultiRouter_CatchAllRoutes(t *testing.T) {
	multi := NewMultiRouter()

	// API with catch-all
	apiRouter := New()
	apiRouter.GET("/files/{filepath...}", dummyHandler)
	multi.Group("/api", apiRouter)

	// Admin with catch-all
	adminRouter := multi.NewGroup("/admin")
	adminRouter.GET("/logs/{logpath...}", dummyHandler)

	// Default with catch-all (no conflict)
	defaultRouter := New()
	defaultRouter.GET("/{path...}", dummyHandler)
	multi.Default(defaultRouter)

	// Should not panic - these don't conflict
}

func TestMultiRouter_DeepNesting(t *testing.T) {
	multi := NewMultiRouter()

	// Create deeply nested route
	apiRouter := New()
	apiRouter.GET("/v1/admin/users/profile", dummyHandler) // /api/v1/admin/users/profile
	multi.Group("/api", apiRouter)

	// Try to create conflicting nested group
	v1Router := New()
	v1Router.GET("/admin/users/profile", dummyHandler) // Would also be /api/v1/admin/users/profile

	defer func() {
		if r := recover(); r != nil {
			msg := r.(string)
			if !strings.Contains(msg, "GROUP CONFLICT") {
				t.Errorf("Expected group conflict error, got: %v", r)
			}
		} else {
			t.Error("Expected panic for deeply nested conflict")
		}
	}()

	multi.Group("/api/v1", v1Router) // Should panic
}

func TestMultiRouter_EmptyGroups(t *testing.T) {
	multi := NewMultiRouter()

	// Create empty groups (should not conflict)
	multi.NewGroup("/api")
	multi.NewGroup("/admin")
	multi.NewGroup("/v1")

	// Add default router
	defaultRouter := New()
	defaultRouter.GET("/home", dummyHandler)
	multi.Default(defaultRouter)

	// Should not panic - empty groups don't conflict
}

func TestMultiRouter_RootGroup(t *testing.T) {
	multi := NewMultiRouter()

	// Create root group
	rootRouter := New()
	rootRouter.GET("/users", dummyHandler)
	multi.Group("/", rootRouter)

	// Try to add default router
	defaultRouter := New()
	defaultRouter.GET("/admin", dummyHandler)
	multi.Default(defaultRouter)

	// Should not panic - root group and default are handled separately
}

func TestMultiRouter_PrefixNormalization(t *testing.T) {
	multi := NewMultiRouter()

	// Test prefix normalization
	apiRouter := New()
	apiRouter.GET("/users", dummyHandler)

	// These should all be normalized to "/api"
	multi.Group("api", apiRouter) // No leading slash

	adminRouter := New()
	adminRouter.GET("/dashboard", dummyHandler)
	multi.Group("/admin/", adminRouter) // Trailing slash should be removed

	// Should not panic - different normalized prefixes
}
