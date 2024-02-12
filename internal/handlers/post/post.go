package post

import (
	"encoding/base64"
	"fmt"
	"test-auth/internal/service/auth"

	"github.com/labstack/echo"
)

type SignInInput struct {
	GUID string `json:"guid" binding:"required"`
}

type RefreshInput struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Auth(ctx echo.Context, service auth.ServiceAuth) (auth.Tokens, error) {
	var inp SignInInput
	if err := ctx.Bind(&inp); err != nil {
		return auth.Tokens{}, err
	}

	res, err := service.SignIn(ctx.Request().Context(), inp.GUID)
	if err != nil {
		return auth.Tokens{}, err
	}

	return res, nil
}

func UserRefresh(ctx echo.Context, service auth.ServiceAuth) (auth.Tokens, error) {
	var inp RefreshInput
	if err := ctx.Bind(&inp); err != nil {
		return auth.Tokens{}, err
	}

	data, err := base64.StdEncoding.DecodeString(inp.RefreshToken)
	if err != nil {
		return auth.Tokens{}, err
	}
	res, err := service.RefreshTokens(ctx.Request().Context(), inp.AccessToken, fmt.Sprintf("%s", data))
	if err != nil {
		return auth.Tokens{}, err
	}

	return res, nil
}
