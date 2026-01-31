package translate

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	deeplAPIKey string
)

const translateAPIURL = "https://api-free.deepl.com/v2/translate"

type translationResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

func Init(apiKey string) {
	deeplAPIKey = apiKey
}

func Get(textToTranslate, sourceLangCode, targetLangCode string) (string, error) {
	apiKey := deeplAPIKey

	data := url.Values{}
	data.Set("source_lang", sourceLangCode)
	data.Set("target_lang", targetLangCode)
	data.Set("text", textToTranslate)

	req, err := http.NewRequest("POST", translateAPIURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "DeepL-Auth-Key "+apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", errors.New("DeepL API Error, status : " + strconv.Itoa(resp.StatusCode))
	}

	var response translationResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.Translations[0].Text, nil
}
