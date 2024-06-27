package portal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	authCookieName = "PortalStudentUserID"
	loginEndpoint  = "http://stda.minia.edu.eg/Portallogin"
	getJCIEndpoint = "http://stda.minia.edu.eg/PortalgetJCI"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnexpectedNil      = errors.New("unexpected nil value")
	ErrMalformedResults   = errors.New("malformed results")
)

type Service struct {
	httpClient *http.Client
}

func NewService() *Service {
	return &Service{
		httpClient: &http.Client{},
	}
}

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

func (p *Service) Login(username, password string) (*http.Cookie, error) {
	req, err := createLoginRequest(username, password)
	if err != nil {
		return nil, fmt.Errorf("error creating login request: %v", err)
	}

	var resp *http.Response
	resp, err = p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer func(resp *http.Response) {
		_ = resp.Body.Close()
	}(resp)

	cookies := resp.Cookies()

	for _, cookie := range cookies {
		if cookie.Name == authCookieName && cookie.Value != "" {
			return cookie, nil
		}
	}

	return nil, ErrInvalidCredentials
}

type StudentResult struct {
	ScopeName string `json:"ScopeName"`
	Year      string `json:"Year"`
	Ds        []struct {
		GradeName        *string `json:"GradeName"`
		Percent          *string `json:"Percent"`
		Total            *string `json:"Total"`
		StudyYearCourses []struct {
			GradeName string `json:"GradeName"`
			Parts     []struct {
				DegreesType    []string `json:"DegreesType"`
				DegreesMax     string   `json:"DegreesMax"`
				Degrees        []string `json:"Degrees"`
				CoursePartName string   `json:"CoursePartName"`
				SemasterName   string   `json:"SemasterName"`
			} `json:"Parts"`
			CourseName  string `json:"CourseName"`
			SuccessFlag string `json:"SuccessFlag"`
			Max         string `json:"Max"`
			Total       string `json:"Total"`
		} `json:"StudyYearCourses"`
	} `json:"ds"`
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

func (p *Service) GetResults(cookie *http.Cookie, uuid string) (*[]StudentResult, error) {
	req, err := createGetResultsRequest(cookie, uuid)

	var resp *http.Response
	resp, err = p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer func(resp *http.Response) {
		_ = resp.Body.Close()
	}(resp)

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var response []StudentResult
	if err = json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &response, nil

}

func GetFirstTranslation(str string) (*string, error) {
	if res := strings.Split(str, "|"); len(res) >= 1 {
		return &res[0], nil
	} else {
		return nil, ErrMalformedResults
	}
}

func FormatResults(results *[]StudentResult) (*string, error) {
	if results == nil {
		return nil, ErrUnexpectedNil
	}

	if len(*results) == 0 {
		msg := "مفيش درجات ولا حاجة لسة يا صديقي!"
		return &msg, nil
	}

	var builder strings.Builder

	for resultsIndex, result := range *results {
		if len(result.Ds) < 1 {
			continue
		}

		if name, err := GetFirstTranslation(result.ScopeName); err == nil {
			builder.WriteString(fmt.Sprintf("⬅ %s (%s)", *name, result.Year))
		} else {
			return nil, ErrMalformedResults
		}

		yearResult := result.Ds[0]

		if yearResult.Percent != nil && yearResult.GradeName != nil && yearResult.Total != nil {
			gradeName, err := GetFirstTranslation(*yearResult.GradeName)
			if err != nil {
				return nil, err
			}

			builder.WriteString(fmt.Sprintf(
				" - (التقدير العام: %s) - (المجموع الكلي: %s) - (النسبة المئوية الكلية: %s%%)",
				*gradeName,
				*yearResult.Total,
				*yearResult.Percent,
			))
		}

		builder.WriteString("\n")

		for courseIndex, course := range yearResult.StudyYearCourses {
			courseName, err := GetFirstTranslation(course.CourseName)
			if err != nil {
				return nil, err
			}

			gradeName, err := GetFirstTranslation(course.GradeName)
			if err != nil {
				return nil, err
			}

			details := fmt.Sprintf(
				"%d) %s - %s - %s/%s",
				courseIndex+1,
				*courseName,
				*gradeName,
				course.Max,
				course.Total,
			)

			builder.WriteString(details)

			if len(course.Parts) > 0 {
				var partsBuilder strings.Builder

				for _, part := range course.Parts {
					if len(part.Degrees) != len(part.DegreesType) {
						continue
					}

					for degreeIndex, degree := range part.Degrees {
						var degreeType *string
						degreeType, err = GetFirstTranslation(part.DegreesType[degreeIndex])
						if err != nil {
							return nil, err
						}

						partsBuilder.WriteString(fmt.Sprintf("%s = %s", *degreeType, degree))

						if degreeIndex != len(part.Degrees)-1 {
							partsBuilder.WriteString(", ")
						}
					}
				}

				partsString := partsBuilder.String()
				if partsString != "" {
					builder.WriteString(fmt.Sprintf(" (%s)", partsString))
				}
			}

			builder.WriteString("\n")
		}

		if resultsIndex != len(*results)-1 {
			builder.WriteString("\n\n\n")
		}
	}

	resultString := builder.String()

	return &resultString, nil
}

type StudentData struct {
	CollageID   string `json:"CollageID"`
	ImagePath   string `json:"ImagePath"`
	UUID        string `json:"UUID"`
	Collage     string `json:"Collage"`
	ScopeUUID   string `json:"ScopeUUID"`
	StdName     string `json:"StdName"`
	Year        string `json:"Year"`
	ShowMessage string `json:"ShowMessage"`
	ID          int64  `json:"ID"`
	StudyYear   string `json:"StudyYear"`
}

func (p *Service) GetStudentData(cookie *http.Cookie) (*StudentData, error) {
	data := url.Values{
		"param0": {"Portal.General"},
		"param1": {"GetStudentPortalData"},
		"param2": {"{\"UserID\":\"\"}"},
	}

	req, err := http.NewRequest("POST", getJCIEndpoint, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.AddCookie(cookie)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	var resp *http.Response
	resp, err = p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer func(resp *http.Response) {
		_ = resp.Body.Close()
	}(resp)

	var body []byte
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	response := &[]StudentData{}
	if err = json.Unmarshal(body, response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &(*response)[0], nil
}
