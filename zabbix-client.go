package zabbix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type API struct {
	url        string
	user       string
	pass       string
	auth       string
	loggedin   bool
	id         int
	httpClient *http.Client
}

type JsonRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Auth    string      `json:"auth,omitempty"`
	ID      int         `json:"id"`
}

type JsonResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   ZbxError    `json:"error"`
	ID      int         `json:"id"`
}

type ZbxError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func (z *ZbxError) Error() string {
	return fmt.Sprintf("Error code: %d, message: %s, data: %s", z.Code, z.Message, z.Data)
}

// NewClient creates new client
func NewClient(url, user, pass string) (*API, error) {
	if url == "" || user == "" || pass == "" {
		return nil, errors.New("you have to provide url, user name and password")
	}
	return &API{url, user, pass, "", false, 0, &http.Client{}}, nil
}

// Login user in Zabbix
func (api *API) Login() (bool, error) {
	params := make(map[string]string)
	params["user"] = api.user
	params["password"] = api.pass

	response, err := api.Request("user.login", params)
	if err != nil {
		return false, err
	}

	// check response error code
	if response.Error.Code != 0 {
		return false, &response.Error
	}

	api.auth = response.Result.(string)
	api.loggedin = true
	return true, nil
}

// Logout user
func (api *API) Logout() (bool, error) {
	emptyParams := make(map[string]string)

	response, err := api.Request("user.logout", emptyParams)
	if err != nil {
		return false, err
	}

	// check response error code
	if response.Error.Code != 0 {
		return false, &response.Error
	}

	api.loggedin = false
	api.auth = ""
	return true, nil
}

// GetHost fetch hosts with given params
func (api *API) GetHost(params interface{}) (JsonResponse, error) {
	response, err := api.Request("host.get", params)
	if err != nil {
		return JsonResponse{}, err
	}
	return response, nil
}

// GetHistory fetch history with given params
func (api *API) GetHistory(params interface{}) (JsonResponse, error) {
	response, err := api.Request("history.get", params)
	if err != nil {
		return JsonResponse{}, err
	}
	return response, nil
}

// ItemUpdate updates items with given params
func (api *API) ItemUpdate(params interface{}) (JsonResponse, error) {
	response, err := api.Request("item.update", params)
	if err != nil {
		return JsonResponse{}, err
	}
	return response, nil
}

// DiscoveryRuleUpdate updates drules with given params
func (api *API) DiscoveryRuleUpdate(params interface{}) (JsonResponse, error) {
	response, err := api.Request("discoveryrule.update", params)
	if err != nil {
		return JsonResponse{}, err
	}
	return response, nil
}

// Request makes request to Zabbix server with given params
func (api *API) Request(method string, params interface{}) (JsonResponse, error) {
	var jsonReq JsonRequest
	jsonReq.Jsonrpc = "2.0"
	jsonReq.Method = method
	jsonReq.Auth = api.auth
	jsonReq.Params = params
	api.id++
	jsonReq.ID = api.id

	encodedReq, err := json.Marshal(jsonReq)
	if err != nil {
		return JsonResponse{}, err
	}

	request, err := http.NewRequest("POST", api.url, bytes.NewBuffer(encodedReq))
	if err != nil {
		return JsonResponse{}, err
	}
	request.Header.Add("Content-type", "application/json-rpc")

	response, err := api.httpClient.Do(request)
	if err != nil {
		return JsonResponse{}, err
	}
	defer response.Body.Close()

	var resultResponse JsonResponse
	var buf bytes.Buffer
	_, err = io.Copy(&buf, response.Body)
	if err != nil {
		return JsonResponse{}, err
	}

	json.Unmarshal(buf.Bytes(), &resultResponse)

	return resultResponse, nil
}
