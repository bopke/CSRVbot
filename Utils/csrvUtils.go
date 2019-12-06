package Utils

import (
	"csrvbot/Config"
	"encoding/json"
	"log"
	"net/http"
)

func GetCSRVCode() (string, error) {
	req, err := http.NewRequest("POST", "https://craftserve.pl/api/generate_voucher", nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("csrvbot", Config.CsrvSecret)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Utils GetCSRVCode Unable to send http request ", err)
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Code string `json:"code"`
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}
	return data.Code, nil
}
