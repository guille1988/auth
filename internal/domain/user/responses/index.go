package responses

import (
	"auth/internal/domain/user/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IndexResponse struct {
	Context *gin.Context
}

func NewIndexResponse(context *gin.Context) *IndexResponse {
	return &IndexResponse{
		Context: context,
	}
}

type indexResponseData struct {
	UUID            string  `json:"uuid"`
	Name            string  `json:"name"`
	Email           string  `json:"email"`
	EmailVerifiedAt *string `json:"email_verified_at"`
	CreatedAt       string  `json:"created_at"`
}

func (response *IndexResponse) Make(users []model.User) {
	collection := make([]indexResponseData, len(users))

	for i, user := range users {
		var emailVerifiedAt *string

		if user.EmailVerifiedAt != nil {
			formatted := user.EmailVerifiedAt.Format("2006-01-02 15:04:05")
			emailVerifiedAt = &formatted
		}

		collection[i] = indexResponseData{
			UUID:            user.UUID.String(),
			Name:            user.Name,
			Email:           user.Email,
			EmailVerifiedAt: emailVerifiedAt,
			CreatedAt:       user.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	response.Context.JSON(http.StatusOK, collection)
}
