package cpanel

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/publicsuffix"
)

func NewWhmApi(hostname string) (*WhmAPI, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return &WhmAPI{}, err
	}

	return &WhmAPI{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint
				},
			},
			Jar: jar,
		},
	}, nil
}

type WhmAPI struct {
	hostname *string
	token    *string
	client   *http.Client
}

func (a *WhmAPI) GenerateURL(endpoint string) (url.URL, error) {
	if a.token == nil {
		return url.URL{}, errors.New("whm api endpoint not activated")
	}
	return url.URL{
		Scheme:   "https",
		Host:     *a.hostname + ":2087",
		Path:     fmt.Sprintf("%s/json-api/%s", *a.token, endpoint),
		RawQuery: "api.version=1",
	}, nil
}

func (a *WhmAPI) Call(
	method string,
	endpoint string,
	args url.Values,
	out interface{},
) error {

	args["api.version"] = []string{"1"}
	url := url.URL{
		Scheme:   "https",
		Host:     *a.hostname + ":2087",
		RawQuery: args.Encode(),
	}
	if a.token != nil {
		url.Path = fmt.Sprintf("%s/json-api/%s", *a.token, endpoint)
	} else {
		url.Path = fmt.Sprintf("/json-api/%s", endpoint)
	}

	request, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}
	response, err := a.client.Do(request)
	if err != nil {
		return err
	}

	err = json.NewDecoder(response.Body).Decode(&out)
	if err != nil {
		return err
	}
	return nil
}

type NumericLimit struct {
	value     int
	unlimited bool
}

func (l *NumericLimit) String() string {
	if l.unlimited {
		return "unlimited"
	}
	return strconv.Itoa(l.value)
}

func (l *NumericLimit) MarshalJSON() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l *NumericLimit) UnmarshalJSON(v []byte) error {
	s := string(v)
	s = strings.Trim(s, "\"")
	if s == "unlimited" {
		l.unlimited = true
		l.value = 0
		return nil
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("value is not a limit: %w", err)
	}
	l.value = i
	l.unlimited = false
	return nil
}
