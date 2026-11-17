package advertisements

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"api-tests-template/internal/constants/path"
	apiRunner "api-tests-template/internal/helpers/api-runner"
)

func HttpGetAdvertisements(t *testing.T, token, textForSearch string) *http.Response {
	return apiRunner.GetRunner().Auth(token).Create().Get(path.AdvertisementsPath).
		Query("search", textForSearch).
		Expect(t).Status(http.StatusOK).
		End().Response
}

func HttpPostAdvertisement(
	t *testing.T,
	token string,
	title string,
	description string,
	price int,
	quantity int,
	photoPaths ...string,
) *http.Response {
	request := apiRunner.GetRunner().
		Auth(token).
		Create().
		Post(path.AdvertisementPath).
		MultipartFormData("title", title).
		MultipartFormData("description", description).
		MultipartFormData("price", strconv.Itoa(price)).
		MultipartFormData("quantity", strconv.Itoa(quantity))

	if len(photoPaths) > 0 {
		request.MultipartFile("photos", photoPaths...)
	}

	return request.
		Expect(t).
		Status(http.StatusCreated).
		End().Response
}

func HttpGetAdvertisement(t *testing.T, token, id string) *http.Response {
	return apiRunner.GetRunner().
		Auth(token).
		Create().
		Get(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		Status(http.StatusOK).
		End().Response
}

func HttpGetAdvertisementPhotos(t *testing.T, token, id string) *http.Response {
	photosPath := fmt.Sprintf(path.AdvertisementPhotosPath, id)

	return apiRunner.GetRunner().
		Auth(token).
		Create().
		Get(photosPath).
		Expect(t).
		Status(http.StatusOK).
		End().Response
}

func HttpDeleteAdvertisement(t *testing.T, token, id string) *http.Response {
	return apiRunner.GetRunner().
		Auth(token).
		Create().
		Delete(path.AdvertisementPath).
		Query("id", id).
		Expect(t).
		Status(http.StatusNoContent).
		End().Response
}

// HttpPostAdvertisementWithStatus создаёт объявление и проверяет ожидаемый HTTP-статус.
func HttpPostAdvertisementWithStatus(
	t *testing.T,
	token, title, description string,
	price, quantity, expectedStatus int,
	photoPaths ...string,
) *http.Response {
	request := apiRunner.GetRunner().
		Auth(token).
		Create().
		Post(path.AdvertisementPath).
		MultipartFormData("title", title).
		MultipartFormData("description", description).
		MultipartFormData("price", strconv.Itoa(price)).
		MultipartFormData("quantity", strconv.Itoa(quantity))

	if len(photoPaths) > 0 {
		request.MultipartFile("photos", photoPaths...)
	}

	return request.Expect(t).Status(expectedStatus).End().Response
}

// HttpPostAdvertisementWithoutRequiredFields отправляет запрос создания без обязательных полей.
func HttpPostAdvertisementWithoutRequiredFields(t *testing.T, token string, expectedStatus int) *http.Response {
	return apiRunner.GetRunner().
		Auth(token).
		Create().
		Post(path.AdvertisementPath).
		MultipartFormData("title", "").
		MultipartFormData("description", "").
		Expect(t).
		Status(expectedStatus).
		End().Response
}
