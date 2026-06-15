# Modular Monolith Refactoring Guide - PocketArtisan Backend

**Date**: June 15, 2026  
**Status**: Approved for Implementation  
**Scope**: Architecture improvements for testability, type safety, and maintainability

---

## Executive Summary

Your backend follows solid modular monolith principles with feature-based vertical slicing. However, two critical architectural issues prevent proper module isolation and testing:

1. **Global shared state** (DB, Redis, JWTService)
2. **Weak type safety** in dependency injection

This guide provides step-by-step refactoring with code examples. Note: Order module is intentionally left unwired and excluded from this refactoring scope.

---

## Architecture Assessment Matrix

| Principle | Current | Target | Impact |
|-----------|---------|--------|--------|
| Module Isolation | ❌ Weak | ✅ Strong | Better testability, easier module swapping |
| Type Safety | ❌ Poor (interface{}) | ✅ Strong | Compile-time checking, IDE support |
| Dependency Clarity | ❌ Hidden (globals) | ✅ Explicit | Visible dependencies, easier debugging |
| Testability | ❌ Difficult | ✅ Easy | Unit tests without container setup |

---

## Part 1: Eliminate Global Variables (P0 - CRITICAL)

### Problem
```go
// config/postgress.go
var DB *gorm.DB  // Global database connection

// config/redis.go
var RDB *redis.Client  // Global Redis client

// config/crypto.go
var instance JWTService  // Global JWT singleton
```

**Issues:**
- All modules implicitly depend on these globals
- Unit testing requires full container setup
- Hidden dependencies not in function signatures
- Violates dependency inversion principle
- Makes code harder to understand (where does DB come from?)

### Solution: Dependency Container Pattern

**Step 1: Create AppContainer struct**

Create `internal/container/container.go`:
```go
package container

import (
	"context"
	"PocketArtisan/internal/modules/auth"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AppContainer holds all application dependencies
type AppContainer struct {
	DB         *gorm.DB
	RDB        *redis.Client
	JWTService auth.JWTService
	// Add other services as needed
	// CacheService CacheService
	// FileStorage  FileStorage
}

// NewAppContainer creates and initializes all dependencies
func NewAppContainer(db *gorm.DB, rdb *redis.Client, jwtService auth.JWTService) *AppContainer {
	return &AppContainer{
		DB:         db,
		RDB:        rdb,
		JWTService: jwtService,
	}
}
```

**Step 2: Update main.go to use container**

Modify `cmd/main.go`:
```go
package main

import (
	"PocketArtisan/config"
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http"
	"PocketArtisan/internal/modules/auth"
)

func main() {
	// Initialize components
	db := config.InitDB()
	rdb := config.InitRedis()
	jwtService := auth.NewJWTService() // Instead of global getter
	
	// Create container
	appContainer := container.NewAppContainer(db, rdb, jwtService)
	
	// Setup router with container
	router := http.SetupRouter(appContainer)
	
	router.Run(":8080")
}
```

**Step 3: Update router to accept container**

Modify `internal/http/router.go`:
```go
package http

import (
	"PocketArtisan/internal/container"
	"PocketArtisan/internal/http/routes"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes all routes with dependency injection
func SetupRouter(appContainer *container.AppContainer) *gin.Engine {
	router := gin.Default()

	// Setup middleware
	SetupMiddleware(router, appContainer)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Register all module routes with container
		routes.RegisterUserRoutes(v1, appContainer)
		routes.RegisterProductRoutes(v1, appContainer)
		routes.RegisterCartRoutes(v1, appContainer)
		routes.RegisterCraftsRoutes(v1, appContainer)
		routes.RegisterCraftsmanApplicationRoutes(v1, appContainer)
		routes.RegisterProductCategoriesRoutes(v1, appContainer)
		routes.RegisterCraftsmanRoutes(v1, appContainer)
		routes.RegisterFileRoutes(v1, appContainer)
	}

	return router
}

func SetupMiddleware(router *gin.Engine, appContainer *container.AppContainer) {
	// Pass container to middleware
	router.Use(middleware.JWTMiddleware(appContainer.JWTService))
	router.Use(middleware.RoleMiddleware())
	router.Use(middleware.CORSMiddleware())
}
```

**Step 4: Update all route registration signatures**

Current (problematic):
```go
func RegisterUserRoutes(router *gin.RouterGroup, db interface{}, rdb interface{}, jwtService interface{}) {
    // ...
}
```

Updated:
```go
func RegisterUserRoutes(router *gin.RouterGroup, appContainer *container.AppContainer) {
    // Access dependencies from container
    db := appContainer.DB
    rdb := appContainer.RDB
    jwtService := appContainer.JWTService
    
    // Register routes...
}
```

**Benefits:**
✅ All dependencies visible in function signature  
✅ Type-safe (no empty interface{})  
✅ Easy to mock in tests  
✅ Single source of truth for dependency initialization  

---

## Part 2: Fix Type Safety (P0 - CRITICAL)

### Problem
```go
func RegisterRoutes(router *gin.RouterGroup, db interface{}, rdb interface{}) {
    // Force cast to expected types - no compile-time checking
    database := db.(*gorm.DB)
    redis := rdb.(*redis.Client)
}
```

**Issues:**
- Passing wrong type causes runtime panic
- IDE can't provide autocomplete
- Compiler can't catch errors
- Refactoring is risky

### Solution: Use Concrete Types

**All route registration functions should use:**
```go
func RegisterOrderRoutes(router *gin.RouterGroup, appContainer *container.AppContainer) {
    // Direct, type-safe access
    db := appContainer.DB           // Type: *gorm.DB (no cast needed)
    rdb := appContainer.RDB         // Type: *redis.Client (no cast needed)
    jwt := appContainer.JWTService  // Type: auth.JWTService (no cast needed)
    
    // Use them directly
    orderGroup := router.Group("/orders")
    orderGroup.POST("", func(c *gin.Context) {
        controller.Create(c, db, rdb, jwt)
    })
}
```

**Benefits:**
✅ Compile-time type checking  
✅ IDE autocomplete works  
✅ Refactoring safer  
✅ Self-documenting code  

---

## Part 3: Define Module Contracts (P1 - HIGH)

### Problem
Modules import from each other's internals without clear contracts:
```go
// ❌ BAD: Importing internal implementation
import "PocketArtisan/internal/modules/product"
CraftsmanID, err := product.GetCraftsmanIDByUsername(...)

// ❌ BAD: Undeclared dependency on Cart in Users module
var userCart entities.Cart
```

### Solution: Service Interfaces as Contracts

**Step 1: Define module interfaces**

Create `internal/modules/product/service.go`:
```go
package product

import "context"

// Service defines the public contract for product module
type Service interface {
    // GetCraftsmanIDByUsername returns the craftsman ID for given username
    GetCraftsmanIDByUsername(ctx context.Context, username string) (string, error)
    
    // GetProductsByCategory returns all products in category
    GetProductsByCategory(ctx context.Context, categoryID string) ([]Product, error)
    
    // GetProductPrice returns current product price
    GetProductPrice(ctx context.Context, productID string) (float64, error)
}

// ProductService implements Service interface
type ProductService struct {
    db *gorm.DB
}

func NewProductService(db *gorm.DB) Service {
    return &ProductService{db: db}
}
```

Create `internal/modules/cart/service.go`:
```go
package cart

import "context"

// Service defines the public contract for cart module
type Service interface {
    // GetUserCart returns cart for given user
    GetUserCart(ctx context.Context, userID string) (*Cart, error)
    
    // AddToCart adds product to user's cart
    AddToCart(ctx context.Context, userID, productID string, quantity int) error
    
    // RemoveFromCart removes product from cart
    RemoveFromCart(ctx context.Context, userID, productID string) error
}

type CartService struct {
    db  *gorm.DB
    rdb *redis.Client
}

func NewCartService(db *gorm.DB, rdb *redis.Client) Service {
    return &CartService{db: db, rdb: rdb}
}
```

**Step 2: Add services to AppContainer**

In `internal/container/container.go`:
```go
package container

import (
	"PocketArtisan/internal/modules/cart"
	"PocketArtisan/internal/modules/product"
	"PocketArtisan/internal/modules/auth"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AppContainer struct {
	DB             *gorm.DB
	RDB            *redis.Client
	JWTService     auth.JWTService
	ProductService product.Service  // NEW
	CartService    cart.Service     // NEW
	// Add other module services here
}

func NewAppContainer(db *gorm.DB, rdb *redis.Client, jwtService auth.JWTService) *AppContainer {
	return &AppContainer{
		DB:             db,
		RDB:            rdb,
		JWTService:     jwtService,
		ProductService: product.NewProductService(db),
		CartService:    cart.NewCartService(db, rdb),
	}
}
```

**Step 3: Update inter-module calls to use contracts**

Before:
```go
// ❌ BAD: Direct internal import and call
import "PocketArtisan/internal/modules/product"
CraftsmanID, err := product.GetCraftsmanIDByUsername(username)
```

After:
```go
// ✅ GOOD: Use service interface through container
craftsmanID, err := uc.productService.GetCraftsmanIDByUsername(ctx, username)
```

**Benefits:**
✅ Clear module boundaries  
✅ Explicit contracts between modules  
✅ Easy to swap implementations  
✅ Mockable for testing  

---

## Part 5: Extract Repository Pattern (P1 - HIGH)

### Problem
Modules directly use GORM throughout:
```go
uc.db.Preload("Items.Product.Images").First(&userCart)
uc.db.Where("craftsman_id = ?", craftsmanID).Find(&products)
```

**Issues:**
- Tightly coupled to GORM ORM
- Hard to swap database implementations
- Query logic scattered across usecases
- Difficult to test without database
- No centralized query logic

### Solution: Repository Abstraction

**Step 1: Create repository interfaces**

Create `internal/repositories/order_repository.go`:
```go
package repositories

import (
	"context"
	"PocketArtisan/internal/entities"
)

// OrderRepository defines data access operations for orders
type OrderRepository interface {
	// Create stores a new order
	Create(ctx context.Context, order *entities.Order) error
	
	// GetByID retrieves order by ID
	GetByID(ctx context.Context, id string) (*entities.Order, error)
	
	// GetAll retrieves all orders
	GetAll(ctx context.Context) ([]entities.Order, error)
	
	// Update modifies existing order
	Update(ctx context.Context, order *entities.Order) error
	
	// Delete removes an order
	Delete(ctx context.Context, id string) error
	
	// GetByUserID retrieves user's orders
	GetByUserID(ctx context.Context, userID string) ([]entities.Order, error)
}

// OrderRepositoryImpl implements OrderRepository using GORM
type OrderRepositoryImpl struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &OrderRepositoryImpl{db: db}
}

func (r *OrderRepositoryImpl) Create(ctx context.Context, order *entities.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *OrderRepositoryImpl) GetByID(ctx context.Context, id string) (*entities.Order, error) {
	var order entities.Order
	err := r.db.WithContext(ctx).First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ... implement other methods ...
```

Create similar repositories for:
- `UserRepository`
- `ProductRepository`
- `CartRepository`
- `CraftsRepository`
- etc.

**Step 2: Update container to include repositories**

In `internal/container/container.go`:
```go
type AppContainer struct {
	DB                 *gorm.DB
	RDB                *redis.Client
	JWTService         auth.JWTService
	
	// Repositories
	OrderRepository    repositories.OrderRepository
	UserRepository     repositories.UserRepository
	ProductRepository  repositories.ProductRepository
	CartRepository     repositories.CartRepository
}

func NewAppContainer(db *gorm.DB, rdb *redis.Client, jwtService auth.JWTService) *AppContainer {
	return &AppContainer{
		DB:                 db,
		RDB:                rdb,
		JWTService:         jwtService,
		OrderRepository:    repositories.NewOrderRepository(db),
		UserRepository:     repositories.NewUserRepository(db),
		ProductRepository:  repositories.NewProductRepository(db),
		CartRepository:     repositories.NewCartRepository(db),
	}
}
```

**Step 3: Update usecases to use repositories**

Before:
```go
func (uc *OrderCreateUseCase) Execute(ctx context.Context, input dto.CreateOrderInput) error {
	var order entities.Order
	uc.db.Preload("Items").First(&order, "user_id = ?", input.UserID)
	uc.db.Create(&newOrder)
}
```

After:
```go
func (uc *OrderCreateUseCase) Execute(ctx context.Context, input dto.CreateOrderInput) error {
	// Type-safe, testable
	order, err := uc.orderRepo.GetByUserID(ctx, input.UserID)
	if err != nil {
		return err
	}
	
	err = uc.orderRepo.Create(ctx, &newOrder)
	return err
}
```

**Benefits:**
✅ Decoupled from ORM implementation  
✅ Easy to mock for tests  
✅ Centralized data access logic  
✅ Can swap PostgreSQL for MySQL without changing business logic  

---

## Part 6: Centralize Error Handling (P1 - HIGH)

### Problem
Error responses are inconsistent across modules:
```go
// Some modules
c.JSON(400, gin.H{"error": "Invalid input"})

// Others
c.JSON(400, map[string]string{"message": "Validation failed"})

// Some with details
c.JSON(500, gin.H{"error": err.Error(), "code": "INTERNAL_SERVER_ERROR"})
```

### Solution: Standardized Error Response

**Create `internal/http/errors.go`:**
```go
package http

import (
	"github.com/gin-gonic/gin"
)

// ErrorResponse standardized error format
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse standardized success format
type SuccessResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
}

// RespondError sends standardized error response
func RespondError(c *gin.Context, statusCode int, code string, message string, details string) {
	c.JSON(statusCode, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}

// RespondSuccess sends standardized success response
func RespondSuccess(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, SuccessResponse{
		Status: "success",
		Data:   data,
	})
}

// Common HTTP errors
func BadRequest(c *gin.Context, message string) {
	RespondError(c, 400, "BAD_REQUEST", message, "")
}

func Unauthorized(c *gin.Context) {
	RespondError(c, 401, "UNAUTHORIZED", "Authentication required", "")
}

func Forbidden(c *gin.Context) {
	RespondError(c, 403, "FORBIDDEN", "Access denied", "")
}

func NotFound(c *gin.Context) {
	RespondError(c, 404, "NOT_FOUND", "Resource not found", "")
}

func InternalError(c *gin.Context, message string) {
	RespondError(c, 500, "INTERNAL_ERROR", "Internal server error", message)
}
```

**Usage in controllers:**
```go
func (cc *CreateController) Handle(c *gin.Context) {
	var input dto.CreateOrderDTO
	
	if err := c.BindJSON(&input); err != nil {
		http.BadRequest(c, "Invalid JSON: "+err.Error())
		return
	}
	
	order, err := cc.usecase.Execute(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, OrderNotFound) {
			http.NotFound(c)
		} else {
			http.InternalError(c, err.Error())
		}
		return
	}
	
	http.RespondSuccess(c, 201, order)
}
```

**Benefits:**
✅ Consistent API responses  
✅ Better client error handling  
✅ Easier debugging  
✅ Professional API design  

---

## Part 7: Move Shared Validation (P1 - HIGH)

### Problem
`users/validator/` should not be module-specific:
```
internal/modules/users/validator/  ❌ WRONG - tied to users module
```

### Solution: Shared Utilities Package

**Reorganize:**
```
internal/validators/              ✅ CORRECT - shared across modules
├── email_validator.go
├── phone_validator.go
├── product_validator.go
├── order_validator.go
└── common.go
```

**Create `internal/validators/validators.go`:**
```go
package validators

import (
	"errors"
	"regexp"
	"strings"
)

// ValidateEmail checks if email is valid
func ValidateEmail(email string) error {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	if !matched {
		return errors.New("invalid email format")
	}
	return nil
}

// ValidateUsername checks if username is valid
func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return errors.New("username must be 3-20 characters")
	}
	if !isAlphanumericUnderscore(username) {
		return errors.New("username can only contain letters, numbers, and underscores")
	}
	return nil
}

// ValidatePhoneNumber checks if phone is valid
func ValidatePhoneNumber(phone string) error {
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	
	if len(cleaned) < 10 {
		return errors.New("invalid phone number")
	}
	return nil
}

func isAlphanumericUnderscore(s string) bool {
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}
```

**Update imports across modules:**
```go
// OLD
import "PocketArtisan/internal/modules/users/validator"
err := validator.ValidateEmail(email)

// NEW
import "PocketArtisan/internal/validators"
err := validators.ValidateEmail(email)
```

**Benefits:**
✅ Reusable across modules  
✅ Single source of truth  
✅ Easier maintenance  

---

## Part 8: Consistent Dependency Injection (P2 - MEDIUM)

### Problem
Mix of global getters and parameter passing:
```go
// Some places
jwtService := auth.GetJWTService()  // Global!

// Others
func RegisterRoutes(..., jwtService auth.JWTService) // Parameter
```

### Solution: Always Use Container

**Rule: NEVER use global getters**

Replace all:
```go
// OLD - Remove these patterns
jwtService := auth.GetJWTService()
database := config.DB
cache := config.RDB

// NEW - Use container exclusively
jwtService := appContainer.JWTService
database := appContainer.DB
cache := appContainer.RDB
```

**In `config/` package - Remove all public getters:**
```go
// Delete these functions
// func GetDB() *gorm.DB
// func GetRedis() *redis.Client
// func GetJWTService() JWTService

// Keep only private initialization
func initDB() *gorm.DB { ... }
func initRedis() *redis.Client { ... }
```

**Benefits:**
✅ All dependencies explicit  
✅ Easier to understand data flow  
✅ Testable without globals  

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [ ] Create `container/container.go`
- [ ] Update `cmd/main.go` to use container
- [ ] Update `http/router.go` signature
- [ ] Fix type safety in route registration

### Phase 2: Core Fix (Week 1-2)
- [ ] Remove global variable getters
- [ ] Update all route functions to use container

### Phase 3: Module Contracts (Week 1-2)
- [ ] Create service interfaces for each module
- [ ] Add services to container
- [ ] Update inter-module calls to use services
- [ ] Update middleware to use container

### Phase 4: Repository Pattern (Week 2-3)
- [ ] Create repository interfaces
- [ ] Implement GORM-based repositories
- [ ] Add repositories to container
- [ ] Update usecases to use repositories

### Phase 5: Polish (Week 3-4)
- [ ] Implement centralized error handling
- [ ] Move validation to shared package
- [ ] Remove all global getters
- [ ] Add tests for dependency injection

---

## Testing Strategy After Refactoring

Once refactored, unit testing becomes trivial:

```go
func TestOrderCreateUseCase(t *testing.T) {
	// Mock repository
	mockOrderRepo := &MockOrderRepository{}
	mockOrderRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	
	// Create usecase with mocks
	usecase := order.NewCreateUseCase(mockOrderRepo, mockProductService)
	
	// Test
	result, err := usecase.Execute(context.Background(), input)
	
	// Verify
	assert.NoError(t, err)
	mockOrderRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}
```

---

## Compliance Checklist

After implementing all changes, verify:

- [ ] No global variable access in modules (all via container)
- [ ] All route functions accept `*container.AppContainer`
- [ ] No `interface{}` type parameters in route registration
- [ ] All order routes registered and working
- [ ] Order transactions handle errors (not just panics)
- [ ] Service interfaces defined for each module
- [ ] All inter-module calls use service interfaces
- [ ] Repository pattern implemented for data access
- [ ] Centralized error handling in use
- [ ] Validation moved to shared package
- [ ] No circular imports
- [ ] No global getters in config package
- [ ] Unit tests can run without Docker container

---

## Architecture After Refactoring

```
┌─────────────────────────────────────┐
│         HTTP Layer (Router)         │
│      (Receives AppContainer)        │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      AppContainer                   │
│  (Single source of dependencies)    │
│  ├─ DB, RDB                         │
│  ├─ Services (interfaces)           │
│  ├─ Repositories (interfaces)       │
│  └─ JWT Service                     │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│      Modules (Use Cases)            │
│  (Receive only needed dependencies) │
│  ├─ No globals                      │
│  ├─ Type-safe injection             │
│  └─ Easily testable                 │
└──────────────┬──────────────────────┘
               │
┌──────────────▼──────────────────────┐
│    Repositories & Services          │
│  (Data access & business logic)     │
│  ├─ Interface-based                 │
│  ├─ Mockable                        │
│  └─ Swappable implementations       │
└─────────────────────────────────────┘
```

---

## Migration Notes

**Do NOT:**
- ❌ Delete config package functions until all modules updated
- ❌ Make breaking changes to entity models
- ❌ Remove routes without replacing them
- ❌ Deploy without testing complete flow

**DO:**
- ✅ Keep old patterns working during migration
- ✅ Test each module after migration
- ✅ Run full integration tests after each phase
- ✅ Deploy incrementally phase by phase

---

## Questions & Clarifications

**Q: Should I implement all at once or gradually?**  
A: Gradually. Implement in phase order (1→5). After Phase 2, the critical issues are fixed. Phases 3-5 improve maintainability.

**Q: Do I need to change the database schema?**  
A: No. This is a purely code-level refactoring. The schema remains unchanged.

**Q: Will this affect API contracts?**  
A: No. External API endpoints remain identical. Only internal structure changes.

**Q: How long will this take?**  
A: ~4-5 weeks working part-time on phases. Phases 1-2 take 1-2 weeks and solve critical issues.

---

## Related Documentation

- [Custom Item Order Flow](./custom-item-order-flow.md)
- [Redis Cache Strategy](./REDIS_CACHE_VERSIONED_WRITE_THROUGH_PATCH.md)
- [Search Optimization](./SEARCH_OPTIMIZATION.md)

---

**Document Version**: 1.0  
**Last Updated**: June 15, 2026  
**Status**: Ready for Implementation
