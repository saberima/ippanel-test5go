package ippanel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"runtime"
)

var (
	ErrUnexpectedResponse = errors.New("The Ippanel API is currently unavailable")
	ErrStatusUnauthorized = errors.New("You api key is not valid")
)

// ListParams ...
type ListParams struct {
	Limit int64 `json:"limit"`
	Page  int64 `json:"page"`
}

// PaginationInfo ...
type PaginationInfo struct {
	Total int64   `json:"total"`
	Limit int64   `json:"limit"`
	Page  int64   `json:"page"`
	Pages int64   `json:"pages"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
}

// BaseResponse base response model
type BaseResponse struct {
	Status       string          `json:"status"`
	Code         ResponseCode    `json:"code"`
	Data         json.RawMessage `json:"data"`
	Meta         *PaginationInfo `json:"meta"`
	ErrorMessage string          `json:"error_message"`
}

// request preform http request
func (sms Ippanel) request(method string, uri string, params map[string]string, data interface{}) (*BaseResponse, error) {
	u := *sms.BaseURL

	// join base url with extra path
	u.Path = path.Join(sms.BaseURL.Path, uri)

	// set query params
	p := url.Values{}
	for key, param := range params {
		p.Add(key, param)
	}
	u.RawQuery = p.Encode()

	marshaledBody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	requestBody := bytes.NewBuffer(marshaledBody)
	req, err := http.NewRequest(method, u.String(), requestBody)
	if err != nil {
		return nil, err
	}

	//req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Apikey", sms.Apikey)
	req.Header.Set("User-Agent", "Ippanel/ApiClient/"+ClientVersion+" Go/"+runtime.Version())

	res, err := sms.Client.Do(req)
	if err != nil || res == nil {
		return nil, err
	}

	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	switch res.StatusCode {
	case http.StatusOK, http.StatusCreated:
		_res := &BaseResponse{}
		if err := json.Unmarshal(responseBody, _res); err != nil {
			return nil, fmt.Errorf("could not decode response JSON, %s: %v", string(responseBody), err)
		}

		return _res, nil
	case http.StatusInternalServerError:
		// Status code 500 is a server error and means nothing can be done at this
		// point.
		return nil, ErrUnexpectedResponse
	case http.StatusUnauthorized:
		// Status code 500 is a server error and means nothing can be done at this
		// point.
		return nil, ErrUnexpectedResponse
	default:
		_res := &BaseResponse{}
		if err := json.Unmarshal(responseBody, _res); err != nil {
			return nil, fmt.Errorf("could not decode response JSON, %s: %v", string(responseBody), err)
		}

		return _res, ParseErrors(_res)
	}
}

// get do get request
func (sms Ippanel) get(uri string, params map[string]string) (*BaseResponse, error) {
	return sms.request("GET", uri, params, nil)
}

// post do post request
func (sms Ippanel) post(uri string, contentType string, data interface{}) (*BaseResponse, error) {
	return sms.request("POST", uri, nil, data)
}
