package handler

import (
	"net/http"
	"strconv"

	"golang-grpc-enterprise-demo/internal/middleware"
	"golang-grpc-enterprise-demo/internal/service"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type RESTHandler struct {
	svc    service.UserService
	jwtMgr *middleware.JWTManager
}

func NewRESTHandler(svc service.UserService, jwtMgr *middleware.JWTManager) *RESTHandler {
	return &RESTHandler{svc: svc, jwtMgr: jwtMgr}
}

func (h *RESTHandler) Register(r *gin.RouterGroup) {
	r.POST("/login", h.Login)
	r.GET("/users/:id", h.GetUser)
	r.POST("/users", h.CreateUser)
	r.GET("/users", h.ListUsers)
}

// LoginRequest represents the login payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"alice@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// TokenResponse represents the login success response.
type TokenResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIs..."`
}

// UserResponse represents a user in API responses.
type UserResponse struct {
	ID        int64  `json:"id" example:"1"`
	Name      string `json:"name" example:"Alice Wang"`
	Email     string `json:"email" example:"alice@example.com"`
	CreatedAt string `json:"created_at" example:"2026-03-02T10:00:00Z"`
}

// CreateUserRequest represents the create-user payload.
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required" example:"Dave Chen"`
	Email    string `json:"email" binding:"required" example:"dave@example.com"`
	Password string `json:"password" binding:"required" example:"secure456"`
}

// ListUsersResponse represents the paginated user list.
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total" example:"3"`
	Page  int            `json:"page" example:"1"`
}

// ErrorResponse represents an error.
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}

// Login godoc
// @Summary      User login
// @Description  Authenticate with email and password, returns JWT token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login credentials"
// @Success      200   {object}  TokenResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      401   {object}  ErrorResponse
// @Router       /api/v1/login [post]
func (h *RESTHandler) Login(c *gin.Context) {
	ctx, span := otel.Tracer("rest").Start(c.Request.Context(), "Login")
	defer span.End()

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}

	user, err := h.svc.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.jwtMgr.Generate(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: token})
}

// GetUser godoc
// @Summary      Get user by ID
// @Description  Returns a single user by their ID
// @Tags         Users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users/{id} [get]
func (h *RESTHandler) GetUser(c *gin.Context) {
	ctx, span := otel.Tracer("rest").Start(c.Request.Context(), "GetUser")
	defer span.End()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.svc.GetByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// CreateUser godoc
// @Summary      Create a new user
// @Description  Register a new user with name, email and password
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUserRequest  true  "User info"
// @Success      201   {object}  UserResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse  "Email already exists"
// @Security     BearerAuth
// @Router       /api/v1/users [post]
func (h *RESTHandler) CreateUser(c *gin.Context) {
	ctx, span := otel.Tracer("rest").Start(c.Request.Context(), "CreateUser")
	defer span.End()

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.Create(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// ListUsers godoc
// @Summary      List users with pagination
// @Description  Returns a paginated list of all users
// @Tags         Users
// @Produce      json
// @Param        page       query     int  false  "Page number"   default(1)
// @Param        page_size  query     int  false  "Items per page" default(10)
// @Success      200        {object}  ListUsersResponse
// @Failure      500        {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /api/v1/users [get]
func (h *RESTHandler) ListUsers(c *gin.Context) {
	ctx, span := otel.Tracer("rest").Start(c.Request.Context(), "ListUsers")
	defer span.End()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	users, total, err := h.svc.List(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]UserResponse, len(users))
	for i, u := range users {
		items[i] = UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, ListUsersResponse{
		Users: items,
		Total: total,
		Page:  page,
	})
}
