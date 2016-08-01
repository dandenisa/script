package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
)

const (
	URL         = "ec2-52-37-174-113.us-west-2.compute.amazonaws.com"
	PROTOCOL    = "http://"
	LOCALHOST   = "localhost"
	PORT_NAME   = "8082"
	AUTHPATH    = "/auth/login"
	ACCOUNTPATH = "/api/accounts"
	PROJECTPATH = "/api/projects"
	USERNAME    = "admin"
	PASSWORD    = "shipyard"
)

func setUrl() string {
	url := PROTOCOL + URL + ":" + PORT_NAME
	return url
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Authentication struct {
	AuthToken string `json:"auth_token"`
}
type ScriptId struct {
	Id string `json:"id"`
}

var token string
var credentials Credentials

func setCredentials(u Credentials) Credentials {
	u.Username = USERNAME
	u.Password = PASSWORD
	return u
}

func postAuthentication() []byte {
	path := setUrl() + AUTHPATH
	c := setCredentials(credentials)
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(c)
	result, err := http.Post(path, "application/json; charset=utf-8", b)

	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(result.Body)
	return body
}

func parseAuthResponse(body []byte) (string, error) {
	var auth Authentication

	err := json.Unmarshal(body, &auth)
	if err != nil {
		fmt.Println(err)
	}
	y := marshalOb(auth)
	split := strings.Split(y, ":")
	authToken := split[1]
	result := authToken[1 : len(authToken)-2]
	return result, err
}

func getIdList(body []byte) []ScriptId {
	var s []ScriptId
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println(err)
	}
	return s

}

/*func getProjectId() {
	body := getProjectsfromApi()
	s := getIdList(body)
	//var p ScriptId
	fmt.Println("response from projects api", string(body))
	for i, data := range s {
		fmt.Println("s[i]:", i, data.Id)
		fmt.Println("s[i] type :", reflect.TypeOf(data.Id))

		//fmt.Println("marshaled id:", i, marshalOb(p[i]))
		i++

	}

}*/
func marshalOb(v interface{}) string {
	vBytes, _ := json.Marshal(v)
	return string(vBytes)
}
func getAuthToken() string {

	body := postAuthentication()
	s, _ := parseAuthResponse(body)
	//fmt.Println("1", string(s))
	x := string(s)
	return x
}

func getAccountsfromApi() {
	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + ACCOUNTPATH
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Access-Token", token)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		_, err := io.Copy(os.Stdout, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func getProjectsfromApi() []byte {
	var body []byte
	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("X-Access-Token", token)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		body, err = ioutil.ReadAll(response.Body)

		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println(string(body))
	return body
}

func getImagesfromApi() {

	var body2 []byte
	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"

	body := getProjectsfromApi()
	s := getIdList(body)
	for i, data := range s {
		projId := url + data.Id + "/images"

		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("images:", i, string(body2))
	}
}
func getTestsFromApi() {
	var body2 []byte

	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"

	body := getProjectsfromApi()
	s := getIdList(body)
	for i, data := range s {
		projId := url + data.Id + "/tests"

		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("test", i, string(body2))
	}

}

/*func getTestId() string {
	body := getTestsFromApi()
	s, _ := parseTestResponse(body)
	fmt.Println("response from projects api", string(body))
	fmt.Println("getProjectId()", string(s))
	x := string(s)
	return x
}*/
func parseTestResponse(body []byte) (string, error) {
	var s []ScriptId
	err := json.Unmarshal(body, &s)
	if err != nil {
		fmt.Println("whoops:", err)
	}
	fmt.Println("after unmarshal:", s)
	y := marshalOb(s)
	final := strings.Split(y, ":")
	final2 := final[1]
	final3 := final2[1 : len(final2)-3]
	fmt.Println("project id:", final3)
	return final3, err
}
func getBuildsFromApi() {
	var body2 []byte

	token = USERNAME + ":" + getAuthToken()

	projectsBody := getProjectsfromApi()
	//testsBody := getTestsFromApi()
	s := getIdList(projectsBody)
	fmt.Println("response from projects api", string(projectsBody))
	for i, data := range s {
		fmt.Println("s[i]:", i, data.Id)
		fmt.Println("s[i] type :", reflect.TypeOf(data.Id))

		projId := "http://ec2-52-37-174-113.us-west-2.compute.amazonaws.com:8082/api/projects/" + data.Id + "/tests" + "builds"
		fmt.Println("images url:", projId)

		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
		}
		fmt.Println("builds:", string(body2))
	}

}

func getRegistriesFromAPi() []byte {
	var body2 []byte

	token = "admin" + ":" + getAuthToken()
	projId := "http://ec2-52-37-174-113.us-west-2.compute.amazonaws.com:8082/api/registries"
	fmt.Println("images url:", projId)
	client := &http.Client{}
	req, err := http.NewRequest("GET", projId, nil)
	req.Header.Add("X-Access-Token", token)
	response, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		defer response.Body.Close()
		body2, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	return body2

}

func main() {

	/*********working methods:**********/
	fmt.Println("post authentication:", string(postAuthentication()))
	fmt.Println("get accounts:")
	getAccountsfromApi()
	fmt.Println("get projects:")
	getProjectsfromApi()
	fmt.Println("get registries:", string(getRegistriesFromAPi()))
	fmt.Println("get images:")
	getImagesfromApi()
	fmt.Println("get tests:")
	getTestsFromApi()
	/*********in progress methods:**********/
	/*fmt.Println("get builds:")
	getBuildsFromApi()*/
}
