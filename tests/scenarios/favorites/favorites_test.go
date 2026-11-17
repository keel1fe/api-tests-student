package favorites

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/tidwall/gjson"

	"api-tests-template/internal/managers/advertisements"
	"api-tests-template/internal/managers/auth"
	authModels "api-tests-template/internal/managers/auth/models"
	favoritesManager "api-tests-template/internal/managers/favorites"

	base "api-tests-template/tests"
)

const (
	favoriteTestDescription = "Тестовое объявление для проверки избранного"
	favoriteTestPrice       = 1000
	favoriteTestQuantity    = 1
	missingAdvertisementID  = "00000000-0000-0000-0000-000000000000"
)

type TestSuite struct {
	suite.Suite
	loginData       authModels.LoginOkResponse
	secondLoginData authModels.LoginOkResponse
	hasSecondUser   bool
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()
	base.Precondition("Авторизация основного тестового пользователя")
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))

	secondLogin, secondPassword := os.Getenv("TEST_LOGIN_SECOND"), os.Getenv("TEST_PASSWORD_SECOND")
	if secondLogin != "" && secondPassword != "" {
		base.Precondition("Авторизация второго пользователя для проверки межпользовательского избранного")
		s.secondLoginData = auth.Login(s.T(), secondLogin, secondPassword)
		s.hasSecondUser = true
	}
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

// requireSecondUser пропускает сценарий, если не настроена вторая тестовая учётная запись.
func (s *TestSuite) requireSecondUser() {
	if !s.hasSecondUser {
		s.T().Skip("для сценариев избранного необходимы TEST_LOGIN_SECOND и TEST_PASSWORD_SECOND — учётная запись другого пользователя")
	}
}

// createAd создаёт объявление с уникальными тестовыми данными.
func (s *TestSuite) createAd(t *testing.T, token, prefix string) modelsAdvertisement {
	t.Helper()
	title := fmt.Sprintf("%s %d", prefix, time.Now().UnixNano())
	photoPath, err := filepath.Abs(filepath.Join("..", "..", "testdata", "test-photo.png"))
	require.NoError(t, err, "не удалось определить путь к тестовой фотографии")
	_, err = os.Stat(photoPath)
	require.NoError(t, err, "тестовая фотография не найдена: %s", photoPath)
	ad := advertisements.CreateAdvertisement(t, token, title, favoriteTestDescription, favoriteTestPrice, favoriteTestQuantity, photoPath)
	return modelsAdvertisement{ID: ad.ID}
}

type modelsAdvertisement struct {
	ID string
}

// favoriteContains проверяет наличие объявления в массиве items ответа GET /favorites.
func favoriteContains(body, advertisementID string) bool {
	items := gjson.Get(body, "items")
	for _, item := range items.Array() {
		if item.Get("advertisement_id").String() == advertisementID ||
			item.Get("ad_id").String() == advertisementID ||
			item.Get("advertisement.id").String() == advertisementID {
			return true
		}
	}
	return false
}

// проверяем добавление, удаление из избранного и удаление самого объявления.
func (s *TestSuite) TestFavoriteLifecycle() {
	s.requireSecondUser()
	t := s.T()
	ad := s.createAd(t, s.secondLoginData.Token, "Favorite lifecycle")
	favoriteAdded := false
	advertisementDeleted := false
	t.Cleanup(func() {
		if favoriteAdded && !advertisementDeleted {
			base.Precondition("Очистка избранного после теста")
			favoritesManager.DeleteFavorite(t, s.loginData.Token, ad.ID, http.StatusNoContent)
		}
		if !advertisementDeleted {
			base.Precondition("Очистка объявления второго пользователя")
			advertisements.DeleteAdvertisement(t, s.secondLoginData.Token, ad.ID)
		}
	})

	base.Precondition("Добавление чужого объявления в избранное")
	favoritesManager.AddFavorite(t, s.loginData.Token, ad.ID, http.StatusCreated)
	favoriteAdded = true

	body := favoritesManager.GetFavorites(t, s.loginData.Token, http.StatusOK)
	require.True(t, favoriteContains(body, ad.ID), "добавленное объявление отсутствует в GET /favorites")

	base.Precondition("Удаление объявления из избранного")
	favoritesManager.DeleteFavorite(t, s.loginData.Token, ad.ID, http.StatusNoContent)
	favoriteAdded = false
	body = favoritesManager.GetFavorites(t, s.loginData.Token, http.StatusOK)
	require.False(t, favoriteContains(body, ad.ID), "удалённое из избранного объявление осталось в GET /favorites")

	base.Precondition("Повторное добавление объявления в избранное перед удалением объявления")
	favoritesManager.AddFavorite(t, s.loginData.Token, ad.ID, http.StatusCreated)
	favoriteAdded = true
	base.Precondition("Удаление самого объявления владельцем")
	advertisements.DeleteAdvertisement(t, s.secondLoginData.Token, ad.ID)
	advertisementDeleted = true
	favoriteAdded = false
	body = favoritesManager.GetFavorites(t, s.loginData.Token, http.StatusOK)
	require.False(t, favoriteContains(body, ad.ID), "удалённое объявление осталось в избранном")
}

// TestFavoriteDuplicate проверяет бизнес-ограничение повторного добавления.
func (s *TestSuite) TestFavoriteDuplicate() {
	s.requireSecondUser()
	t := s.T()
	ad := s.createAd(t, s.secondLoginData.Token, "Favorite duplicate")
	favoriteAdded := false
	t.Cleanup(func() {
		if favoriteAdded {
			favoritesManager.DeleteFavorite(t, s.loginData.Token, ad.ID, http.StatusNoContent)
		}
		advertisements.DeleteAdvertisement(t, s.secondLoginData.Token, ad.ID)
	})

	favoritesManager.AddFavorite(t, s.loginData.Token, ad.ID, http.StatusCreated)
	favoriteAdded = true
	body := favoritesManager.AddFavorite(t, s.loginData.Token, ad.ID, http.StatusConflict)
	require.NotEmpty(t, gjson.Get(body, "message").String(), "для повторного добавления отсутствует сообщение об ошибке")
}

// TestFavoriteOwnAdvertisement проверяет запрет добавления собственного объявления.
func (s *TestSuite) TestFavoriteOwnAdvertisement() {
	t := s.T()
	base.Precondition("Создание собственного объявления для проверки бизнес-ограничения избранного")
	ad := s.createAd(t, s.loginData.Token, "Own favorite")
	t.Cleanup(func() {
		advertisements.DeleteAdvertisement(t, s.loginData.Token, ad.ID)
	})
	body := favoritesManager.AddFavorite(t, s.loginData.Token, ad.ID, http.StatusForbidden)
	require.NotEmpty(t, gjson.Get(body, "message").String(), "для собственного объявления отсутствует сообщение об ошибке")
}

// TestFavoriteNonexistentAdvertisement проверяет добавление несуществующего объявления.
func (s *TestSuite) TestFavoriteNonexistentAdvertisement() {
	base.Precondition("Добавление несуществующего объявления в избранное")
	body := favoritesManager.AddFavorite(s.T(), s.loginData.Token, missingAdvertisementID, http.StatusBadRequest)
	require.NotEmpty(s.T(), gjson.Get(body, "message").String(), "для несуществующего объявления отсутствует сообщение об ошибке")
}

// TestDeleteNonexistentFavorite проверяет удаление отсутствующего в избранном объявления.
func (s *TestSuite) TestDeleteNonexistentFavorite() {
	base.Precondition("Удаление несуществующего объявления из избранного")
	body := favoritesManager.DeleteFavorite(s.T(), s.loginData.Token, missingAdvertisementID, http.StatusNotFound)
	require.NotEmpty(s.T(), gjson.Get(body, "message").String(), "для удаления несуществующего избранного отсутствует сообщение об ошибке")
}

// проверяем GET, POST и DELETE без токена.
func (s *TestSuite) TestFavoritesUnauthorized() {
	t := s.T()
	base.Precondition("Проверка всех ручек избранного без Authorization")
	getBody := favoritesManager.GetFavorites(t, "", http.StatusUnauthorized)
	require.Equal(t, "unauthorized", gjson.Get(getBody, "error").String())
	require.Equal(t, "Authorization header required", gjson.Get(getBody, "message").String())

	postBody := favoritesManager.AddFavorite(t, "", missingAdvertisementID, http.StatusUnauthorized)
	require.Equal(t, "unauthorized", gjson.Get(postBody, "error").String())
	require.Equal(t, "Authorization header required", gjson.Get(postBody, "message").String())

	deleteBody := favoritesManager.DeleteFavorite(t, "", missingAdvertisementID, http.StatusUnauthorized)
	require.Equal(t, "unauthorized", gjson.Get(deleteBody, "error").String())
	require.Equal(t, "Authorization header required", gjson.Get(deleteBody, "message").String())
}

// TestFavoritesIncorrectToken проверяет GET, POST и DELETE избранного с неверным токеном.
func (s *TestSuite) TestFavoritesIncorrectToken() {
	t := s.T()
	base.Precondition("Проверка всех ручек избранного с некорректным токеном")
	body := favoritesManager.GetFavorites(t, "incorrect_token", http.StatusUnauthorized)
	require.Equal(t, "unauthorized", gjson.Get(body, "error").String())
	require.Equal(t, "Invalid or expired token", gjson.Get(body, "message").String())

	body = favoritesManager.AddFavorite(t, "incorrect_token", missingAdvertisementID, http.StatusUnauthorized)
	require.Equal(t, "unauthorized", gjson.Get(body, "error").String())
	require.Equal(t, "Invalid or expired token", gjson.Get(body, "message").String())

	body = favoritesManager.DeleteFavorite(t, "incorrect_token", missingAdvertisementID, http.StatusUnauthorized)
	require.Equal(t, "unauthorized", gjson.Get(body, "error").String())
	require.Equal(t, "Invalid or expired token", gjson.Get(body, "message").String())
}
