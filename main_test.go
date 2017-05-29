// main_test.go

package main_test

import (
    "os"
    "log"
    "bytes"
    "net/http"
    "encoding/json"
    "net/http/httptest"
    "testing"

    "github.com/kelvins/goapi"
)

var a main.App

func TestMain(m *testing.M) {
    a = main.App{}
    a.Initialize("root", "", "rest_api_example")

    ensureTableExists()

    code := m.Run()

    clearTable()

    os.Exit(code)
}

func ensureTableExists() {
    if _, err := a.DB.Exec(tableCreationQuery); err != nil {
        log.Fatal(err)
    }
}

func clearTable() {
    a.DB.Exec("DELETE FROM users")
    a.DB.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
}

const tableCreationQuery = `
CREATE TABLE IF NOT EXISTS users
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    age INT NOT NULL
)`

func TestEmptyTable(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/users", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    if body := response.Body.String(); body != "[]" {
        t.Errorf("Expected an empty array. Got %s", body)
    }
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
    rr := httptest.NewRecorder()
    a.Router.ServeHTTP(rr, req)

    return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d. Got %d\n", expected, actual)
    }
}

func TestGetNonExistentUser(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/user/45", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusNotFound, response.Code)

    var m map[string]string
    json.Unmarshal(response.Body.Bytes(), &m)
    if m["error"] != "User not found" {
        t.Errorf("Expected the 'error' key of the response to be set to 'User not found'. Got '%s'", m["error"])
    }
}

func TestCreateUser(t *testing.T) {
    clearTable()

    payload := []byte(`{"name":"test user","30`)

    req, _ := http.NewRequest("POST", "/user", bytes.NewBuffer(payload))
    response := executeRequest(req)

    checkResponseCode(t, http.StatusCreated, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["name"] != "test user" {
        t.Errorf("Expected user name to be 'test user'. Got '%v'", m["name"])
    }

    if m["age"] != 30 {
        t.Errorf("Expected user age to be '30'. Got '%v'", m["price"])
    }

    // the id is compared to 1.0 because JSON unmarshaling converts numbers to
    // floats, when the target is a map[string]interface{}
    if m["id"] != 1.0 {
        t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
    }
}

func addUsers(count int) {
    if count < 1 {
        count = 1
    }

    for i := 0; i < count; i++ {
        a.DB.Exec("INSERT INTO users(name, age) VALUES($1, $2)", "User "+strconv.Itoa(i), (i*10))
    }
}

func TestGetUser(t *testing.T) {
    clearTable()
    addUsers(1)

    req, _ := http.NewRequest("GET", "/user/1", nil)
    response := executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdateProduct(t *testing.T) {
    clearTable()
    addUsers(1)

    req, _ := http.NewRequest("GET", "/user/1", nil)
    response := executeRequest(req)
    var originalUser map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &originalUser)

    payload := []byte(`{"name":"test user - updated name","age":21}`)

    req, _ = http.NewRequest("PUT", "/user/1", bytes.NewBuffer(payload))
    response = executeRequest(req)

    checkResponseCode(t, http.StatusOK, response.Code)

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)

    if m["id"] != originalProduct["id"] {
        t.Errorf("Expected the id to remain the same (%v). Got %v", originalUser["id"], m["id"])
    }

    if m["name"] == originalProduct["name"] {
        t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalUser["name"], m["name"], m["name"])
    }

    if m["age"] == originalProduct["age"] {
        t.Errorf("Expected the age to change from '%v' to '%v'. Got '%v'", originalUser["age"], m["age"], m["age"])
    }
}
