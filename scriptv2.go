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
type SenderStatistics struct {
	Username   string
	Images     []ScriptImageDetails
	Accounts   []ScriptAccounts
	Projects   []ScriptProjects
	Builds     []ScriptBuilds
	Registries []ScriptRegistries
	Tests      []ScriptTests
	Results    []ScriptBuildResults
}

type ScriptBuilds struct {
	Id        string
	ProjectId string
	TestId    string
	StartTime string
	Status    Status
}
type Status struct {
	Status string
}
type Results struct {
	ResultEntries []string
}
type ScriptProjects struct {
	Name         string
	Author       string
	CreationTime string
	LastRunTime  string
	Status       string
}

type ScriptAccounts struct {
	Id        string   `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Roles     []string `json:"roles"`
}

type ScriptImageDetails struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	ImageId     string `json:"imageId"`
	Description string `json:"description"`
	Status      string `json:"status"`
	RegistryId  string `json:"registryId"`

	Location       string `json:"location"`
	SkipImageBuild string `json:"skipImageBuild"`

	ProjectId string `json:"projectId"`
}

type ScriptRegistries struct {
	Name string
	Addr string
}
type ScriptTests struct {
	Id        string `json:"id"`
	ProjectId string `json:"projectId"`
	Provider  Provider
}
type Provider struct {
	providerType string
}
type ScriptBuildResults struct {
	ID            string
	BuildId       string
	ResultEntries []string
}

var projectInfo []ScriptProjects
var registryInfo []ScriptRegistries
var accountInfo []ScriptAccounts
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
func getAccountsfromApi() []byte {
	var body []byte

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
		body, err = ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	return body
}

func unmarshalAccounts(body []byte) []ScriptAccounts {
	error := json.Unmarshal(body, &accountInfo)
	if error != nil {
		fmt.Println("error:", error)
	}
	return accountInfo
}
func unmarshalRegistries(body []byte) []ScriptRegistries {
	error := json.Unmarshal(body, &registryInfo)
	if error != nil {
		fmt.Println("error:", error)
	}
	return registryInfo
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
	return body
}

func getImagesfromApi() []ScriptImageDetails {
	var result []ScriptImageDetails
	var body2 []byte
	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"

	body := getProjectsfromApi()
	s := getIdList(body)
	for _, data := range s {
		projId := url + data.Id + "/images"

		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptImageDetails{}
			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	/*for i := 0; i < len(result); i++ {

		fmt.Printf("this is the result: %d,%v\n", i, result[i])
	}*/
	return result
}
func getTestsFromApi() []ScriptTests {
	var body2 []byte
	var result []ScriptTests

	token = USERNAME + ":" + getAuthToken()
	url := setUrl() + PROJECTPATH + "/"

	body := getProjectsfromApi()
	s := getIdList(body)
	for _, data := range s {
		projId := url + data.Id + "/tests"

		client := &http.Client{}
		req, err := http.NewRequest("GET", projId, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptTests{}

			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	return result
}

func getBuildsFromApi() []ScriptBuilds {
	var body2 []byte
	var result []ScriptBuilds

	token = USERNAME + ":" + getAuthToken()

	testsBody := getTestsFromApi()
	for _, data := range testsBody {
		testId := data.Id
		projId := data.ProjectId

		id := "http://ec2-52-37-174-113.us-west-2.compute.amazonaws.com:8082/api/projects/" + projId + "/tests/" + testId + "/builds"
		client := &http.Client{}
		req, err := http.NewRequest("GET", id, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptBuilds{}

			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
		}
	}
	return result

}

func getResultsFromApi() []ScriptBuildResults {
	var body2 []byte
	var result []ScriptBuildResults

	token = USERNAME + ":" + getAuthToken()

	buildsBody := getBuildsFromApi()
	for _, data := range buildsBody {
		testId := data.TestId
		projId := data.ProjectId
		buildId := data.Id
		id := "http://ec2-52-37-174-113.us-west-2.compute.amazonaws.com:8082/api/projects/" + projId + "/tests/" + testId + "/builds/" + buildId + "/results"
		client := &http.Client{}
		req, err := http.NewRequest("GET", id, nil)
		req.Header.Add("X-Access-Token", token)
		response, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		} else {
			myResult := []ScriptBuildResults{}

			defer response.Body.Close()
			body2, err = ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}
			json.Unmarshal(body2, &myResult)
			result = append(result, myResult...)
		}
	}
	return result

}

func getRegistriesFromAPi() []byte {
	var body2 []byte

	token = "admin" + ":" + getAuthToken()
	projId := "http://ec2-52-37-174-113.us-west-2.compute.amazonaws.com:8082/api/registries"
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

func unmarshalProjects(body []byte) []ScriptProjects {
	error := json.Unmarshal(body, &projectInfo)
	if error != nil {
		fmt.Println("error:", error)
	}
	return projectInfo
}

func setStatistics(stats SenderStatistics) SenderStatistics {
	uname := "admin"
	acc := unmarshalAccounts(getAccountsfromApi())
	img := getImagesfromApi()
	proj := unmarshalProjects(getProjectsfromApi())
	tst := getTestsFromApi()
	reg := unmarshalRegistries(getRegistriesFromAPi())
	bld := getBuildsFromApi()
	res := getResultsFromApi()
	stats.Username = uname
	stats.Accounts = acc
	stats.Images = img
	stats.Projects = proj
	stats.Tests = tst
	stats.Registries = reg
	stats.Builds = bld
	stats.Results = res
	return stats
}

func postResponse() {
	var stats SenderStatistics
	s := setStatistics(stats)
	b := new(bytes.Buffer)

	json.NewEncoder(b).Encode(s)
	res1, _ := http.Post("https://httpbin.org/post", "application/json; charset=utf-8", b)
	io.Copy(os.Stdout, res1.Body)
}

type Countries []ScriptBuildResults

func main() {
	postResponse()
}
