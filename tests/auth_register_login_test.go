package tests

import (
	"fmt"
	"testing"
	"time"

	authService "go-sso/internal/services/auth"
	"go-sso/tests/suite"

	"github.com/brianvoe/gofakeit"
	"github.com/golang-jwt/jwt/v5"
	gossov1 "github.com/passwordhash/protos/gen/go/go-sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	emptyAppID = 0
	appID      = 1
	// appSecret должен совпадать с тем, что используется в tests.migrations
	appSecret = "test-secret"

	passDefaultLen = 10

	errMsgDuplicateRegistration = "user already exists"
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &gossov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserUuid())

	respLogin, err := st.AuthClient.Login(ctx, &gossov1.LoginRequest{
		Email:    email,
		Password: pass,
		AppId:    appID,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	loginTime := time.Now()

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	fmt.Println(claims)
	assert.Equal(t, respReg.GetUserUuid(), claims["uuid"].(string))
	assert.Equal(t, email, claims["email"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	// check if exp of token is in correct range, ttl get from st.Cfg.TokenTTL
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_DuplicateRegistratioin(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &gossov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserUuid())

	respReg, err = st.AuthClient.Register(ctx, &gossov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserUuid())
	assert.ErrorContains(t, err, authService.ErrUserExists.Error())
}

func TestRegisterLogin_InvalidEmail(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &gossov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserUuid())

	respReg, err = st.AuthClient.Register(ctx, &gossov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserUuid())
	assert.ErrorContains(t, err, authService.ErrUserExists.Error())
}

func TestRegisterLogin_InvalidPassword(t *testing.T) {
	ctx, st := suite.New(t)

	email := gofakeit.Email()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &gossov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserUuid())

	respReg, err = st.AuthClient.Register(ctx, &gossov1.RegisterRequest{
		Email:    email,
		Password: pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserUuid())
	assert.ErrorContains(t, err, authService.ErrUserExists.Error())
}

func TestGetSigningKey(t *testing.T) {
	ctx, st := suite.New(t)

	// Проверяем, что при пустом имени приложения возвращается ошибка
	invalidArgResp, err := st.AuthClient.SigningKey(ctx, &gossov1.SigningKeyRequest{
		AppName: "",
	})
	require.Error(t, err)
	assert.Empty(t, invalidArgResp.GetSigningKey())

	appName := gofakeit.BuzzWord()
	// Первый запрос должен вернуть должен сгенерировать новый ключ
	firstResp, err := st.AuthClient.SigningKey(ctx, &gossov1.SigningKeyRequest{
		AppName: appName,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, firstResp.GetSigningKey())

	// Повторой запрос должен вернуть тот же ключ
	secondResp, err := st.AuthClient.SigningKey(ctx, &gossov1.SigningKeyRequest{
		AppName: appName,
	})
	require.NoError(t, err)
	assert.Equal(t, firstResp.GetSigningKey(), secondResp.GetSigningKey())
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
