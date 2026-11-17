package advertisements

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	advertisementsClient "api-tests-template/internal/client/http/advertisements"
	"api-tests-template/internal/managers/advertisements/models"
)

// получаем список объявлений по поисковому запросу.
func GetAdvertisementsList(t *testing.T, token, text string) models.AdvertisementsList {
	t.Helper()
	raw := advertisementsClient.HttpGetAdvertisements(t, token, text)
	defer raw.Body.Close()
	body, err := io.ReadAll(raw.Body)
	require.NoError(t, err, "не удалось прочитать список объявлений")
	var response models.AdvertisementsList
	require.NoError(t, json.Unmarshal(body, &response), "не удалось разобрать список объявлений")
	return response
}

// создаем объявление и возвращаем его данные.
func CreateAdvertisement(t *testing.T, token, title, description string, price, quantity int, photoPaths ...string) models.Advertisement {
	t.Helper()
	raw := advertisementsClient.HttpPostAdvertisement(t, token, title, description, price, quantity, photoPaths...)
	defer raw.Body.Close()
	body, err := io.ReadAll(raw.Body)
	require.NoError(t, err, "не удалось прочитать ответ при создании объявления")
	var response models.Advertisement
	require.NoError(t, json.Unmarshal(body, &response), "не удалось разобрать ответ создания объявления: %s", string(body))
	require.NotEmpty(t, response.ID, "API вернул пустой ID объявления")
	return response
}

// GetAdvertisement получает объявление по ID.
func GetAdvertisement(t *testing.T, token, id string) models.Advertisement {
	t.Helper()
	raw := advertisementsClient.HttpGetAdvertisement(t, token, id)
	defer raw.Body.Close()
	body, err := io.ReadAll(raw.Body)
	require.NoError(t, err, "не удалось прочитать объявление")
	var response models.Advertisement
	require.NoError(t, json.Unmarshal(body, &response), "не удалось разобрать объявление")
	return response
}

// получаем фотографии объявления.
func GetAdvertisementPhotos(t *testing.T, token, id string) []models.Photo {
	t.Helper()
	raw := advertisementsClient.HttpGetAdvertisementPhotos(t, token, id)
	defer raw.Body.Close()
	body, err := io.ReadAll(raw.Body)
	require.NoError(t, err, "не удалось прочитать фотографии")
	var response []models.Photo
	require.NoError(t, json.Unmarshal(body, &response), "не удалось разобрать фотографии")
	return response
}

// удаляем объявление по ID.
func DeleteAdvertisement(t *testing.T, token, id string) {
	t.Helper()
	raw := advertisementsClient.HttpDeleteAdvertisement(t, token, id)
	defer raw.Body.Close()
}

// CreateAdvertisementInvalid отправляет запрос с пустыми обязательными полями и возвращает ошибку API.
func CreateAdvertisementInvalid(t *testing.T, token string, expectedStatus int) string {
	t.Helper()
	raw := advertisementsClient.HttpPostAdvertisementWithoutRequiredFields(t, token, expectedStatus)
	defer raw.Body.Close()
	body, err := io.ReadAll(raw.Body)
	require.NoError(t, err, "не удалось прочитать ошибку создания объявления")
	return string(body)
}
