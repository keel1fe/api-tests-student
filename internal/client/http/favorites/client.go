package favorites

import (
	"net/http"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

type addFavoriteRequest struct {
	AdvertisementID string `json:"advertisement_id"`
}

// HttpGetFavorites получает список избранного с ожидаемым статусом.
func HttpGetFavorites(t *testing.T, token string, expectedStatus int) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().
		Get(path.FavoritesPath).
		Expect(t).Status(expectedStatus).End().Response
}

// HttpPostFavorite добавляет объявление в избранное с ожидаемым статусом.
func HttpPostFavorite(t *testing.T, token, advertisementID string, expectedStatus int) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().
		Post(path.FavoritesPath).
		ContentType("application/json").
		JSON(addFavoriteRequest{AdvertisementID: advertisementID}).
		Expect(t).Status(expectedStatus).End().Response
}

// HttpDeleteFavorite удаляет объявление из избранного с ожидаемым статусом.
func HttpDeleteFavorite(t *testing.T, token, advertisementID string, expectedStatus int) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().
		Delete(path.FavoritesPath + "/" + advertisementID).
		Expect(t).Status(expectedStatus).End().Response
}
