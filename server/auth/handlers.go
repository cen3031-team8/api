package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/cen3031-team8/api/database"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const (
	AdminUserID   int64  = 1
	AdminUsername string = "admin"
	JWTSecret     string = "cen3031-auth-secret"
)

type AuthService struct {
	queries *database.Queries
	log     *zap.Logger
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type ProfileResponse struct {
	UserID    int64                    `json:"userId"`
	Username  string                   `json:"username"`
	Email     string                   `json:"email"`
	Inventory []database.UserInventory `json:"inventory"`
	IsAdmin   bool                     `json:"isAdmin"`
}

type AddPokemonRequest struct {
	TargetUserID int64 `json:"targetUserId"` // Optional, for admin override
}

type PokemonResponse struct {
	Item     string `json:"item"`
	Quantity int32  `json:"quantity"`
}

type UpdateInventoryRequest struct {
	Item     string `json:"item" binding:"required"`
	Quantity int32  `json:"quantity" binding:"required"`
	UserID   int64  `json:"userId"` // For admin updates
}

type CustomClaims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthService(queries *database.Queries, log *zap.Logger) *AuthService {
	return &AuthService{
		queries: queries,
		log:     log,
	}
}

func (as *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (as *AuthService) VerifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (as *AuthService) GenerateJWT(userID int64, username string) (string, error) {
	claims := CustomClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(JWTSecret))
	return tokenString, err
}

func (as *AuthService) VerifyJWT(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (as *AuthService) IsAdmin(userID int64, username string) bool {
	return userID == AdminUserID || username == AdminUsername
}

func (as *AuthService) HandleRegister(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check if user already exists
	_, err := as.queries.GetUserByUsername(c.Request.Context(), req.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	if err != sql.ErrNoRows {
		as.log.Error("Database error checking username", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Hash password
	passwordHash, err := as.HashPassword(req.Password)
	if err != nil {
		as.log.Error("Password hashing error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	// Create user
	user, err := as.queries.CreateUser(c.Request.Context(), database.CreateUserParams{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		as.log.Error("User creation error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User creation failed"})
		return
	}

	// Generate JWT
	token, err := as.GenerateJWT(user.UserID, user.Username)
	if err != nil {
		as.log.Error("JWT generation error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Token:    token,
	})
}

func (as *AuthService) HandleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Get user by username
	user, err := as.queries.GetUserByUsername(c.Request.Context(), req.Username)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		as.log.Error("Database error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify password
	if !as.VerifyPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT
	token, err := as.GenerateJWT(user.UserID, user.Username)
	if err != nil {
		as.log.Error("JWT generation error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		UserID:   user.UserID,
		Username: user.Username,
		Email:    user.Email,
		Token:    token,
	})
}

func (as *AuthService) HandleProfile(c *gin.Context) {
	claims, err := as.extractClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user info
	user, err := as.queries.GetUserByID(c.Request.Context(), claims.UserID)
	if err != nil {
		as.log.Error("Error fetching user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		return
	}

	// Get inventory
	inventory, err := as.queries.GetUserInventory(c.Request.Context(), claims.UserID)
	if err != nil {
		as.log.Error("Error fetching inventory", zap.Error(err))
		inventory = []database.UserInventory{}
	}

	c.JSON(http.StatusOK, ProfileResponse{
		UserID:    user.UserID,
		Username:  user.Username,
		Email:     user.Email,
		Inventory: inventory,
		IsAdmin:   as.IsAdmin(claims.UserID, claims.Username),
	})
}

func (as *AuthService) HandleAddPokemon(c *gin.Context) {
	claims, err := as.extractClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req AddPokemonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default to current user if no targetUserId provided
		req.TargetUserID = claims.UserID
	}

	// If targeting a different user, verify admin
	if req.TargetUserID != claims.UserID && !as.IsAdmin(claims.UserID, claims.Username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can modify other players' inventory"})
		return
	}

	// Add or update enemy pokemon to inventory
	err = as.queries.AddOrUpdateInventoryItem(c.Request.Context(), database.AddOrUpdateInventoryItemParams{
		UserID:   req.TargetUserID,
		Item:     "enemy",
		Quantity: 1,
	})
	if err != nil {
		as.log.Error("Error updating inventory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add pokemon"})
		return
	}

	// Return updated inventory item
	item, err := as.queries.GetUserInventoryItem(c.Request.Context(), database.GetUserInventoryItemParams{
		UserID: req.TargetUserID,
		Item:   "enemy",
	})
	if err != nil {
		as.log.Error("Error fetching inventory item", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated pokemon"})
		return
	}

	c.JSON(http.StatusOK, PokemonResponse{
		Item:     item.Item,
		Quantity: item.Quantity,
	})
}

func (as *AuthService) HandleUpdateInventory(c *gin.Context) {
	claims, err := as.extractClaims(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Default to current user
	targetUserID := claims.UserID
	if req.UserID != 0 {
		targetUserID = req.UserID
	}

	// If targeting a different user, verify admin
	if targetUserID != claims.UserID && !as.IsAdmin(claims.UserID, claims.Username) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can modify other players' inventory"})
		return
	}

	// Update inventory
	err = as.queries.UpdateInventoryItem(c.Request.Context(), database.UpdateInventoryItemParams{
		UserID:   targetUserID,
		Item:     req.Item,
		Quantity: req.Quantity,
	})
	if err != nil {
		as.log.Error("Error updating inventory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Inventory updated"})
}

func (as *AuthService) extractClaims(c *gin.Context) (*CustomClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing authorization header")
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return nil, errors.New("invalid authorization header")
	}

	return as.VerifyJWT(tokenString)
}
