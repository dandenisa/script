package main

import (
	"encoding/json"
	"fmt"
	r "github.com/dancannon/gorethink"
	"github.com/shipyard/shipyard/model"

	"bytes"
	"github.com/shipyard/shipyard/utils/auth"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

const (
	accountsTable   = "accounts"
	projectsTable   = "projects"
	imagesTable     = "images"
	testsTable      = "tests"
	buildsTable     = "builds"
	registriesTable = "registries"
	rethinkDbPort   = "28015"
	database        = "shipyard"
)

var session *r.Session
var imageInfo ScriptImageDetails
var projectInfo ScriptProjects
var registryInfo ScriptRegistries
var accountInfo ScriptAccounts
var testInfo ScriptTests
var buildInfo ScriptBuilds
var statistics SenderStatistics

type SenderStatistics struct {
	Username   string
	Images     []ScriptImageDetails
	Accounts   []ScriptAccounts
	Projects   []ScriptProjects
	Builds     []ScriptBuilds
	Registries []ScriptRegistries
	Tests      []ScriptTests
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
	Roles    []string
	Tokens   Tokens
	Username string
}
type Tokens struct {
	UserAgent string
}
type ScriptImageDetails struct {
	Id          string
	ProjectId   string
	Name        string
	Description string
	Location    string
}

type ScriptRegistries struct {
	Name string
	Addr string
}
type ScriptTests struct {
	ProjectId string
	Provider  Provider
}
type Provider struct {
	providerType string
}

func getContainerIp() string {
	command := "docker inspect --format '{{ .NetworkSettings.Networks.shipyard_default.IPAddress }}' shipyard_rethinkdb_1"
	result, _ := exec.Command("/bin/sh", "-c", command).Output()
	ip := strings.TrimSpace(string(result))
	return ip
}

func initDatabse() {
	var err error
	ip := getContainerIp()
	addr := ip + ":" + rethinkDbPort

	session, err = r.Connect(r.ConnectOpts{
		Address:  addr,
		Database: database,
	})

	if err != nil {
		fmt.Println(err)
		return
	}
}

func retrieveAllImages() []ScriptImageDetails {
	rows, err := r.Table(imagesTable).Run(session)
	if err != nil {
		fmt.Println(err)
		return statistics.Images
	}
	images := []*model.Image{}
	err2 := rows.All(&images)
	if err2 != nil {
		fmt.Println(err2)
		return statistics.Images
	}
	for _, p := range images {
		x := unmarshalImage(p)
		statistics.Images = append(statistics.Images, x)
	}
	return statistics.Images
}
func retrieveAllAccounts() []ScriptAccounts {
	rows, err := r.Table(accountsTable).Run(session)
	if err != nil {
		fmt.Println(err)
		return statistics.Accounts
	}
	accounts := []*auth.Account{}
	err2 := rows.All(&accounts)
	if err2 != nil {
		fmt.Println(err2)
		return statistics.Accounts
	}
	for _, p := range accounts {
		statistics.Accounts = append(statistics.Accounts, unmarshalAccount(p))
	}
	return statistics.Accounts
}

func retrieveAllProjects() []ScriptProjects {
	rows, err := r.Table(projectsTable).Run(session)
	if err != nil {
		fmt.Println(err)
		return statistics.Projects
	}
	projects := []*model.Project{}
	err2 := rows.All(&projects)
	if err2 != nil {
		fmt.Println(err2)
		return statistics.Projects
	}
	for _, p := range projects {
		statistics.Projects = append(statistics.Projects, unmarshalProject(p))
	}
	return statistics.Projects
}

func retrieveAllBuilds() []ScriptBuilds {
	rows, err := r.Table(buildsTable).Run(session)
	if err != nil {
		fmt.Println(err)
		return statistics.Builds
	}
	builds := []*model.Build{}
	err2 := rows.All(&builds)
	if err2 != nil {
		fmt.Println(err2)
		return statistics.Builds
	}
	for _, p := range builds {

		statistics.Builds = append(statistics.Builds, unmarshalBuild(p))
	}
	return statistics.Builds
}
func retrieveAllRegistries() []ScriptRegistries {
	rows, err := r.Table(registriesTable).Run(session)
	if err != nil {
		fmt.Println(err)
		return statistics.Registries

	}
	registries := []*model.Registry{}
	err2 := rows.All(&registries)
	if err2 != nil {
		fmt.Println(err2)
		return statistics.Registries

	}
	for _, p := range registries {

		statistics.Registries = append(statistics.Registries, unmarshalRegistry(p))
	}
	return statistics.Registries
}
func retrieveAllTests() []ScriptTests {
	rows, err := r.Table(testsTable).Run(session)
	if err != nil {
		fmt.Println(err)
		return statistics.Tests

	}
	tests := []*model.Test{}
	err2 := rows.All(&tests)
	if err2 != nil {
		fmt.Println(err2)
		return statistics.Tests
	}
	for _, p := range tests {
		statistics.Tests = append(statistics.Tests, unmarshalTest(p))

	}
	return statistics.Tests

}

func countObjects(tableName string) {
	cursor, err := r.Table(tableName).Count().Run(session)
	if err != nil {
		fmt.Println(err)
		return
	}

	var cnt int
	cursor.One(&cnt)
	cursor.Close()
	fmt.Print("Number of ", tableName, ": ")
	printMarshaledObject(marshalObject(cnt))

}

func unmarshalImage(v interface{}) ScriptImageDetails {
	marshaledBytes := marshalObject(v)
	unmarshaledBytes := json.Unmarshal(marshaledBytes, &imageInfo)
	if unmarshaledBytes != nil {
		panic(unmarshaledBytes)
	}
	return imageInfo
}

func unmarshalRegistry(v interface{}) ScriptRegistries {
	marshaledBytes := marshalObject(v)
	unmarshaledBytes := json.Unmarshal(marshaledBytes, &registryInfo)
	if unmarshaledBytes != nil {
		panic(unmarshaledBytes)
	}
	return registryInfo
}

func unmarshalProject(v interface{}) ScriptProjects {
	marshaledBytes := marshalObject(v)
	unmarshaledBytes := json.Unmarshal(marshaledBytes, &projectInfo)
	if unmarshaledBytes != nil {
		panic(unmarshaledBytes)
	}
	return projectInfo
}

func unmarshalAccount(v interface{}) ScriptAccounts {
	marshaledBytes := marshalObject(v)
	unmarshaledBytes := json.Unmarshal(marshaledBytes, &accountInfo)
	if unmarshaledBytes != nil {
		panic(unmarshaledBytes)
	}
	return accountInfo
}

func unmarshalBuild(v interface{}) ScriptBuilds {
	marshaledBytes := marshalObject(v)
	unmarshaledBytes := json.Unmarshal(marshaledBytes, &buildInfo)
	if unmarshaledBytes != nil {
		panic(unmarshaledBytes)
	}
	return buildInfo
}

func unmarshalTest(v interface{}) ScriptTests {
	marshaledBytes := marshalObject(v)
	unmarshaledBytes := json.Unmarshal(marshaledBytes, &testInfo)
	if unmarshaledBytes != nil {
		panic(unmarshaledBytes)
	}

	return testInfo
}

func marshalObject(v interface{}) []byte {
	vBytes, _ := json.Marshal(v)
	return vBytes
}

func printMarshaledObject(vBytes []byte) {
	fmt.Println(string(vBytes))
}

func CreateStatistics(stats SenderStatistics) SenderStatistics {
	uname := "admin"
	img := retrieveAllImages()
	reg := retrieveAllRegistries()
	bld := retrieveAllBuilds()
	prj := retrieveAllProjects()
	acc := retrieveAllAccounts()
	tst := retrieveAllTests()
	stats.Username = uname
	stats.Images = img
	stats.Registries = reg
	stats.Builds = bld
	stats.Projects = prj
	stats.Accounts = acc
	stats.Tests = tst
	return stats
}

func main() {
	initDatabse()

	var stats SenderStatistics
	s := CreateStatistics(stats)
	b := new(bytes.Buffer)

	json.NewEncoder(b).Encode(s)
	res, _ := http.Post("https://httpbin.org/post", "application/json; charset=utf-8", b)
	io.Copy(os.Stdout, res.Body)

}
