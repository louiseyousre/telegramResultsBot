package portal

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

func createLoginRequest(username, password string) (*http.Request, error) {
	data := url.Values{
		"UserName": {username},
		"Password": {password},
	}

	req, err := http.NewRequest("POST", loginEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	return req, nil
}

func createGetResultsRequest(cookie *http.Cookie, uuid string) (*http.Request, error) {
	data := url.Values{
		"param0": {"Portal.Results"},
		"param1": {"GetAllResults"},
		"param2": {fmt.Sprintf("{\"UUID\":\"%s\"}", uuid)},
	}

	req, err := http.NewRequest("POST", getJCIEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.AddCookie(cookie)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	return req, nil
}

func createGetStudentDataRequest(cookie *http.Cookie) (*http.Request, error) {
	data := url.Values{
		"param0": {"Portal.General"},
		"param1": {"GetStudentPortalData"},
		"param2": {"{\"UserID\":\"\"}"},
	}

	req, err := http.NewRequest("POST", getJCIEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.AddCookie(cookie)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	return req, nil
}
