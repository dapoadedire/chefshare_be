package api
import (

	"net/http"
	"github.com/dapoadedire/chefshare_be/store"
	"github.com/gin-gonic/gin"
)


type registeredUserRequest struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Bio            string `json:"bio"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
}




type UserHandler struct {
	UserStore store.UserStore
}
func NewUserHandler(userStore store.UserStore) *UserHandler {
	return &UserHandler{
		UserStore: userStore,
	}
}
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req registeredUserRequest
	err := c.ShouldBindJSON(&req); 
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &store.User{
		Username:       req.Username,
		Email:          req.Email,
		Bio:            req.Bio,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		ProfilePicture: req.ProfilePicture,
	}
	err = user.PasswordHash.SetPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set password"})
		return
	}

	err = h.UserStore.CreateUser(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}