package main

import (
	"net/http"
	"os"
	"encoding/xml"
	"testing"
	"net/http/httptest"
	"bytes"
	"io"
	"fmt"
	"sort"
	"strings"
	"strconv"
	"encoding/json"
	"errors"
	"time"
)

// код писать тут

type Root struct {
	Row []UserInfo `xml:"row"`
}

type UserInfo struct {
	Id        int    `xml:"id" json:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Name      string `json:"Name"`
	Age       int    `xml:"age" json:"age"`
	About     string `xml:"about" json:"about"`
	Gender    string `xml:"gender" json:"gender"`
}



func TestSearchClient_BadSearchRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	cases := []SearchRequest{
		SearchRequest{
			Limit:      2,
			Offset:     2,
			OrderField: "some",
			OrderBy:    0,
		},
		SearchRequest{
			Limit:      2,
			Offset:     2,
			OrderField: "Age",
			OrderBy:    -20,
		},
		SearchRequest{
			Limit:      -1,
			Offset:     2,
			Query:      "Boyd",
			OrderField: "",
			OrderBy:    0,
		},
		SearchRequest{
			Limit:      100500,
			Offset:     2,
			OrderField: "Age",
			OrderBy:    0,
		},
		SearchRequest{
			Limit:      2,
			Offset:     -1,
			Query:      "Boyd",
			OrderField: "Age",
			OrderBy:    0,
		},
	}

	for _, item := range cases {
		search := &SearchClient{
			AccessToken: "secret",
			URL:         ts.URL,
		}

		search.FindUsers(item)
	}
}

func TestSearchClient_BadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	badUrl := &SearchClient{
		AccessToken: "secret",
		URL:         "error",
	}

	badToken := &SearchClient{
		AccessToken: "badToken",
		URL:         ts.URL,
	}

	badUrl.FindUsers(SearchRequest{})
	badToken.FindUsers(SearchRequest{})

}

func TestSearchClient_Timeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Timeout))
	defer ts.Close()

	search := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}

	search.FindUsers(SearchRequest{})

}

func TestSearchClient_FatalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(FatalError))
	defer ts.Close()

	search := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}

	search.FindUsers(SearchRequest{})
}

func TestSearchClient_Json(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(Json))
	defer ts.Close()

	search := &SearchClient{
		AccessToken: "secret",
		URL:         ts.URL,
	}

	search.FindUsers(SearchRequest{Query: "badRequest"})
	search.FindUsers(SearchRequest{Query: "empty"})
	search.FindUsers(SearchRequest{})
}

func Json(w http.ResponseWriter, r *http.Request) {
	var badJson[]byte
	var userList []User
	query := r.URL.Query().Get("query")

	if query == "badRequest" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(badJson)
	} else if query == "empty" {
		empty, _ := json.Marshal(userList)
		w.Write(empty)
	} else {
		w.Write(badJson)
	}
}

func FatalError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

func Timeout(w http.ResponseWriter, r *http.Request) {
	timer := time.NewTimer(time.Second * 1)
	<-timer.C
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	var userList []UserInfo

	token := r.Header.Get("AccessToken")

	if token != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userList = getData().Row

	sortField := r.URL.Query().Get("order_field")
	orderBy := r.URL.Query().Get("order_by")
	err := sortByField(sortField, orderBy, userList)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		if err.Error() == "ErrorBadOrderField" {
			w.Write([]byte(`{"Error":"ErrorBadOrderField"}`))
		} else {
			w.Write([]byte(`{"Error":"UnknownBadRequest"}`))
		}
		return
	}

	query := r.URL.Query().Get("query")

	if query != "" {
		userList = findUser(query, userList)
	}

	offset := r.URL.Query().Get("offset")
	limit := r.URL.Query().Get("limit")

	if offset != "" && limit != "" {
		userList = getUserRange(userList, offset, limit)
	}

	if r.Method == http.MethodGet {
		response, _ := json.Marshal(userList)
		w.Write(response)
	}
}

func getUserRange(userList []UserInfo, offsetStr string, limitStr string) []UserInfo {
	offset, _ := strconv.Atoi(offsetStr)
	limit, _ := strconv.Atoi(limitStr)

	return userList[offset : limit+offset]
}

func getData() Root {
	var root Root

	file, err := os.Open("dataset.xml")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	buffer := new(bytes.Buffer)
	io.Copy(buffer, file)
	err = xml.Unmarshal(buffer.Bytes(), &root)

	if err != nil {
		fmt.Println("err", err)
		panic(err)
	}

	addName(root.Row)

	return root
}

func addName(userList []UserInfo) {
	for i := 0; i < len(userList); i++ {
		user := &userList[i]
		user.Name = user.FirstName + " " + user.LastName
	}
}

func findUser(query string, userList []UserInfo) []UserInfo {
	var result []UserInfo

	for _, user := range userList {
		if strings.Contains(user.Name, query) || strings.Contains(user.Name, query) {
			result = append(result, user)
		}
	}

	return result
}

func sortByField(field string, order string, userList []UserInfo) error {

	if order != "0" && order != "1" && order != "-1" {
		return errors.New("")
	}

	if field == "" || field == "Name" {
		switch order {
		case "1":
			sort.Slice(userList, func(i, j int) bool {
				return userList[i].Name < userList[j].Name
			})
		case "-1":
			sort.Slice(userList, func(i, j int) bool {
				return userList[i].Name > userList[j].Name
			})
		}

		return nil
	}

	if field == "Age" {
		switch order {
		case "1":
			sort.Slice(userList, func(i, j int) bool {
				return userList[i].Age < userList[j].Age
			})
		case "-1":
			sort.Slice(userList, func(i, j int) bool {
				return userList[i].Age > userList[j].Age
			})
		}

		return nil
	}

	if field == "Id" {
		switch order {
		case "1":
			sort.Slice(userList, func(i, j int) bool {
				return userList[i].Id < userList[j].Id
			})
		case "-1":
			sort.Slice(userList, func(i, j int) bool {
				return userList[i].Id > userList[j].Id
			})
		}
		return nil
	}

	return errors.New("ErrorBadOrderField")
}
