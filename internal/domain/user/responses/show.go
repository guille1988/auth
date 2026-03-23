package responses

import (
	"auth/internal/domain/user/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ShowResponse struct {
	Context *gin.Context
}

func NewShowResponse(context *gin.Context) *ShowResponse {
	return &ShowResponse{
		Context: context,
	}
}

type showResponseData struct {
	UUID            string  `json:"uuid"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	EmailVerifiedAt *string `json:"email_verified_at"`
	CreatedAt       string  `json:"created_at"`
}

func (response *ShowResponse) Make(user *model.User) {
	var emailVerifiedAt *string

	if user.EmailVerifiedAt != nil {
		formatted := user.EmailVerifiedAt.Format("2006-01-02 15:04:05")
		emailVerifiedAt = &formatted
	}

	response.Context.JSON(http.StatusOK,
		showResponseData{
			UUID:            user.UUID.String(),
			Name:            user.Name,
			Email:           user.Email,
			EmailVerifiedAt: emailVerifiedAt,
			CreatedAt:       user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	)
}
