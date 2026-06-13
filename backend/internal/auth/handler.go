package auth

import (
	"errors"
	"net/http"

	"github.com/cinema-booking/backend/internal/repository"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	verifier   *FirebaseVerifier
	jwtService *JWTService
	users      *repository.UserRepository
}

func NewHandler(verifier *FirebaseVerifier, jwtService *JWTService, users *repository.UserRepository) *Handler {
	return &Handler{
		verifier:   verifier,
		jwtService: jwtService,
		users:      users,
	}
}

type loginRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

type userResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

type loginResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token is required"})
		return
	}

	identity, err := h.verifier.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		if errors.Is(err, ErrInvalidFirebaseToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid firebase token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify firebase token"})
		return
	}

	user, err := h.users.UpsertFromFirebase(c.Request.Context(), identity.UID, identity.Email, identity.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upsert user"})
		return
	}

	accessToken, err := h.jwtService.Sign(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{
		AccessToken: accessToken,
		User: userResponse{
			UserID: user.ID.Hex(),
			Email:  user.Email,
			Name:   user.Name,
			Role:   user.Role,
		},
	})
}

func (h *Handler) Me(c *gin.Context) {
	userIDHex := c.GetString(ContextUserIDKey)
	userID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.users.FindByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, userResponse{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
	})
}
