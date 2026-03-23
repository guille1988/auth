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
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (response *ShowResponse) Make(user *model.User) {
	response.Context.JSON(http.StatusOK,
		showResponseData{
			UUID:      user.UUID.String(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	)
}
