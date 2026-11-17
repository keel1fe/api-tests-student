package favorites

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	favoritesClient "api-tests-template/internal/client/http/favorites"
)

// возвращаем тело ответа ручки избранного.
func GetFavorites(t *testing.T, token string, expectedStatus int) string {
	t.Helper()
	response := favoritesClient.HttpGetFavorites(t, token, expectedStatus)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err, "не удалось прочитать ответ избранного")
	return string(body)
}

// добавляем объявление в избранное и возвращает тело ответа.
func AddFavorite(t *testing.T, token, advertisementID string, expectedStatus int) string {
	t.Helper()
	response := favoritesClient.HttpPostFavorite(t, token, advertisementID, expectedStatus)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err, "не удалось прочитать ответ добавления в избранное")
	return string(body)
}

// удаляем объявление из избранного и возвращает тело ответа.
func DeleteFavorite(t *testing.T, token, advertisementID string, expectedStatus int) string {
	t.Helper()
	response := favoritesClient.HttpDeleteFavorite(t, token, advertisementID, expectedStatus)
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err, "не удалось прочитать ответ удаления из избранного")
	return string(body)
}
