package utils

import (
	"encoding/json"
	"fmt"
	"github.com/rtmelsov/GopherMart/internal/models"
	"io"
	"net/http"
)

func GetAccrual(url string, num int64) (*models.Accural, *models.Error) {
	var order *models.Accural
	resp, err := http.Get(fmt.Sprintf("%s/api/orders/%v", url, num))
	if err != nil {
		return nil, &models.Error{
			Error: err.Error(),
			Code:  http.StatusInternalServerError,
		}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	err = json.Unmarshal(body, order)
	if err != nil {
		return nil, &models.Error{
			Error: err.Error(),
			Code:  http.StatusInternalServerError,
		}
	}

	return order, nil
}
