package advertisementCreate

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
	"api-tests-template/internal/managers/auth/models"

	base "api-tests-template/tests"
)

const (
	testPrice    = 1500
	testQuantity = 3
	testTitle    = "QA test advertisement"
	testPhoto    = "test-photo.png"
)

type TestSuite struct {
	suite.Suite
	loginData models.LoginOkResponse
}

func TestSuiteRun(t *testing.T) {
	suite.Run(t, &TestSuite{})
}

func (s *TestSuite) SetupSuite() {
	base.SetupSuite()
	base.Precondition("Авторизация тестового пользователя")
	s.loginData = auth.Login(s.T(), os.Getenv("TEST_LOGIN"), os.Getenv("TEST_PASSWORD"))
}

func (s *TestSuite) TearDownSuite() {
	base.TearDownSuite()
}

// проверяем создание объявления со всеми параметрами.
func (s *TestSuite) TestCreateAdvertisementPositive() {
	t := s.T()
	title := fmt.Sprintf("%s %d", testTitle, time.Now().UnixNano())
	description := "Тестовое объявление для проверки API"
	photoPath, err := filepath.Abs(filepath.Join("..", "..", "testdata", testPhoto))
	require.NoError(t, err, "не удалось определить путь к тестовой фотографии")
	_, err = os.Stat(photoPath)
	require.NoError(t, err, "тестовая фотография не найдена: %s", photoPath)

	base.Precondition("Создание объявления с уникальным заголовком, описанием, ценой, количеством и фотографией")
	created := advertisements.CreateAdvertisement(t, s.loginData.Token, title, description, testPrice, testQuantity, photoPath)
	t.Cleanup(func() {
		base.Precondition("Удаление созданного тестового объявления")
		advertisements.DeleteAdvertisement(t, s.loginData.Token, created.ID)
	})

	s.Run("GET по ID возвращает переданные параметры", func() {
		tt := s.T()
		base.Precondition("Получение объявления по ID и сверка всех переданных полей")
		found := advertisements.GetAdvertisement(tt, s.loginData.Token, created.ID)
		require.Equal(tt, created.ID, found.ID)
		require.Equal(tt, title, found.Title)
		require.Equal(tt, description, found.Description)
		require.Equal(tt, testPrice, found.Price)
		require.Equal(tt, testQuantity, found.Quantity)
	})

	s.Run("Объявление находится по полному названию", func() {
		tt := s.T()
		base.Precondition("Поиск объявления по полному вхождению названия")
		result := advertisements.GetAdvertisementsList(tt, s.loginData.Token, title)
		found := false
		for _, item := range result.Items {
			if item.ID == created.ID && item.Title == title {
				found = true
				break
			}
		}
		require.True(tt, found, "созданное объявление отсутствует в результатах поиска")
	})

	s.Run("Фотография доступна через ручку фотографий", func() {
		tt := s.T()
		base.Precondition("Получение загруженной фотографии объявления")
		photos := advertisements.GetAdvertisementPhotos(tt, s.loginData.Token, created.ID)
		require.NotEmpty(tt, photos, "API не вернул фотографии созданного объявления")
		require.NotEmpty(tt, photos[0].URL, "у загруженной фотографии отсутствует URL")
	})
}

// проверяем отклонение запроса с пустыми обязательными полями.
func (s *TestSuite) TestCreateAdvertisementInvalidData() {
	base.Precondition("Отправка объявления без обязательных title и description")
	body := advertisements.CreateAdvertisementInvalid(s.T(), s.loginData.Token, http.StatusBadRequest)
	require.NotEmpty(s.T(), gjson.Get(body, "error").String(), "в ответе отсутствует код ошибки")
	require.NotEmpty(s.T(), gjson.Get(body, "message").String(), "в ответе отсутствует описание ошибки")
}

// провеярем запрет создания объявления без авторизации.
func (s *TestSuite) TestCreateAdvertisementUnauthorized() {
	base.Precondition("Попытка создания объявления с отсутствующим токеном")
	body := advertisements.CreateAdvertisementInvalid(s.T(), "", http.StatusUnauthorized)
	require.Equal(s.T(), "unauthorized", gjson.Get(body, "error").String())
	require.Equal(s.T(), "Authorization header required", gjson.Get(body, "message").String())
}

// проверяем запрет создания объявления с неверным токеном.
func (s *TestSuite) TestCreateAdvertisementIncorrectToken() {
	base.Precondition("Попытка создания объявления с неверным токеном")
	body := advertisements.CreateAdvertisementInvalid(s.T(), "incorrect_token", http.StatusUnauthorized)
	require.Equal(s.T(), "unauthorized", gjson.Get(body, "error").String())
	require.Equal(s.T(), "Invalid or expired token", gjson.Get(body, "message").String())
}
