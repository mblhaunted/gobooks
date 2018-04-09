package main

import (
    "bytes"
    "fmt"
    "encoding/json"
    "math/rand"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

// Basic test READ test
func TestGetBooksBasic(t *testing.T) {
    expectedCode := 200
    req, _ := http.NewRequest("GET", "/books", nil)
    resp := httptest.NewRecorder()
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    r.ServeHTTP(resp, req)
    checkResponseCode(expectedCode, resp.Code, t)
    var respJson []map[string]interface{}
    err := json.NewDecoder(resp.Body).Decode(&respJson)
    if err != nil {
        t.Errorf("bad json response: %s", resp.Body.String())
    }
}

// Basic tests for bad URIs
func TestGetBooksBadUris(t *testing.T) {
    var badUris []string
    expectedCode := 404
    badUris = append(badUris, "[][][][]")
    badUris = append(badUris, "/*")
    badUris = append(badUris, "/\n")
    badUris = append(badUris, "/all books")
    badUris = append(badUris, "/get/1/")
    badUris = append(badUris, "/1/get")
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    for _, element := range badUris {
        resp := httptest.NewRecorder()
        req, _ := http.NewRequest(
            "GET",
            fmt.Sprintf("/books%v", element),
            nil)
        r.ServeHTTP(resp, req)
        checkResponseCode(expectedCode, resp.Code, t)
    }
}

// Create a new Book and verify response.
func TestCreateBookBasic(t *testing.T) {
    testTitle := "A Tale of Two Code Challenges"
    testDate := "2010-Dec-26"
    expectedCode := 200
    book := Book{Title: testTitle, PublishDate: testDate}
    resp := createBook(book)
    checkResponseCode(expectedCode, resp.Code, t)
    var respJson map[string]interface{}
    err := json.NewDecoder(resp.Body).Decode(&respJson)
    if err != nil {
        t.Errorf("bad json response: %s", resp.Body.String())
    }
    if respJson["title"] != testTitle {
        t.Errorf("expected title %s, was: %s", testTitle, respJson["title"])
    }
}

// Attempt to create a book with bad date
// We only accept shortdate, eg; 2001-Sep-26
func TestCreateBookBadValues(t *testing.T) {
    var badDates []string
    var badRatings []int
    badDates = append(badDates, "2012-12-26")
    badDates = append(badDates, "2012-September-30")
    badDates = append(badDates, "December 21st, 2001")
    badDates = append(badDates, "{\"json\": \"structure\"}")
    badDates = append(badDates, "2001-Jan-35")
    badDates = append(badDates, "2012/06/16")
    badDates = append(badDates, "10/Jan/2008")
    badRatings = append(badRatings, -2)
    badRatings = append(badRatings, 20)
    badRatings = append(badRatings, 4)
    badRatings = append(badRatings, 203003)
    badRatings = append(badRatings, -49849)
    badRatings = append(badRatings, 9999999)
    expectedCode := 400
    var respJson map[string]interface{}
    for index, element := range badDates {
        postJson := Book{Title: fmt.Sprintf("The Future %d", index), PublishDate: element}
        resp := createBook(postJson)
        checkResponseCode(expectedCode, resp.Code, t)
        err := json.NewDecoder(resp.Body).Decode(&respJson)
        if err != nil {
            t.Errorf("bad json response: %s", resp.Body.String())
        }
        if respJson["publish_date"] != element {
            t.Errorf("expected %s, was: %s", element, respJson["publish_date"])
        }
    }
    for index, element := range badRatings {
        postJson := Book{Title: fmt.Sprintf("The Past %d", index), Rating: element}
        resp := createBook(postJson)
        checkResponseCode(expectedCode, resp.Code, t)
        err := json.NewDecoder(resp.Body).Decode(&respJson)
        if err != nil {
            t.Errorf("bad json response: %s", resp.Body.String())
        }
        if respJson["rating"] != float64(element) {
            t.Errorf("expected %d, was: %d", element, respJson["rating"])
        }
    }
}

func TestCreateBookStringsTooLong(t *testing.T) {
    testDate := "2001-Jan-20"
    testTitle := "A Tale of Two Code Challenges"
    testAuthor := "Charles Dickens"
    badTitle := getRandomString(256)
    badAuthor := getRandomString(1000)
    badPublisher := getRandomString(10000)
    expectedCode := 400
    postJson := Book{Title: badTitle, PublishDate: testDate}
    resp := createBook(postJson)
    checkResponseCode(expectedCode, resp.Code, t)
    postJson = Book{Title: testTitle, Author: badAuthor, PublishDate: testDate}
    resp = createBook(postJson)
    checkResponseCode(expectedCode, resp.Code, t)
    postJson = Book{
        Title: testTitle,
        Author: testAuthor,
        PublishDate: testDate,
        Publisher: badPublisher}
    resp = createBook(postJson)
    checkResponseCode(expectedCode, resp.Code, t)
}

// Attempt to create a book with bad data
func TestCreateBookBadDataRaw(t *testing.T) {
    var badDatum [][]byte
    badDatum = append(badDatum, []byte("{[jks1\nkdk"))
    badDatum = append(badDatum, []byte("{[}"))
    badDatum = append(badDatum, []byte("{\"title\": \"foo\"}"))
    badDatum = append(badDatum, []byte("{\"title\": \"A-Bar\", \"author\": \""))
    badDatum = append(badDatum, []byte("{}"))
    expectedCode := 400
    for _, element :=  range badDatum {
        req, _ := http.NewRequest("POST", "/books", bytes.NewReader(element))
        resp := httptest.NewRecorder()
        db := GetDatabaseSession()
        defer db.Close()
        r := GetApiEngine(db)
        r.ServeHTTP(resp, req)
        checkResponseCode(expectedCode, resp.Code, t)
        var respJson map[string]interface{}
        err := json.NewDecoder(resp.Body).Decode(&respJson)
        if err != nil {
            t.Errorf("bad json response: %s", resp.Body.String())
        }
    }
}

// Create a new Book, update it, verify
func TestUpdateBookBasic(t *testing.T) {
    testTitle := "crUdz"
    testAuthor := "Ghost"
    testRating := 3
    testDate := "2001-Dec-01"
    expectedCode := 200
    book := Book{
        Title: testTitle,
        Author: testAuthor,
        Rating: testRating,
        PublishDate: testDate}
    resp := createBook(book)
    checkResponseCode(expectedCode, resp.Code, t)
    var respJson map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&respJson)
    if respJson["checked_out"].(bool) {
        t.Errorf("this book shouldn't be checked out")
    }
    bookId := int(respJson["ID"].(float64) + 0.5)
    // update book
    checkedOutBook := Book{Title: "Beef", CheckedOut: true}
    resp = updateBook(bookId, checkedOutBook)
    checkResponseCode(expectedCode, resp.Code, t)
    _ = json.NewDecoder(resp.Body).Decode(&respJson)
    if !respJson["checked_out"].(bool) {
        t.Errorf("udpated book remains not checked out")
    }
    // fetch and verify
    resp = getBook(bookId)
    checkResponseCode(expectedCode, resp.Code, t)
    _ = json.NewDecoder(resp.Body).Decode(&respJson)
    if !respJson["checked_out"].(bool) {
        t.Errorf("udpated book remains not checked out on GET query")
    }
    if respJson["title"].(string) != "Beef" {
        t.Errorf("udpated book remains not checked out on GET query")
    }
}

// Try to update a book with bad values
func TestUpdateBookBadValues(t *testing.T) {
    testTitle := "crUdz"
    testAuthor := "Ghost"
    testRating := 3
    testDate := "2001-Dec-01"
    expectedCode := 200
    book := Book{
        Title: testTitle,
        Author: testAuthor,
        Rating: testRating,
        PublishDate: testDate}
    resp := createBook(book)
    checkResponseCode(expectedCode, resp.Code, t)
    var respJson map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&respJson)
    // try to update with a few bad values
    expectedCode = 400
    bookId := int(respJson["ID"].(float64) + 0.5)
    badRatingBook := Book{Title: "Vacation", Rating: 50000000}
    resp = updateBook(bookId, badRatingBook)
    checkResponseCode(expectedCode, resp.Code, t)
    badRatingBook = Book{Title: "Vouchers", Rating: -50000000}
    resp = updateBook(bookId, badRatingBook)
    checkResponseCode(expectedCode, resp.Code, t)
    badDateBook := Book{Title: "BringDollars", PublishDate: "10/10/2001"}
    resp = updateBook(bookId, badDateBook)
    checkResponseCode(expectedCode, resp.Code, t)
}

// Create a book, then delete it
func TestDeleteBook(t *testing.T) {
    testTitle := "Not Long For This Position"
    testAuthor := "David Dennison"
    testRating := 3
    testDate := "2016-Jan-01"
    expectedCode := 200
    book := Book{
        Title: testTitle,
        Author: testAuthor,
        Rating: testRating,
        PublishDate: testDate}
    resp := createBook(book)
    checkResponseCode(expectedCode, resp.Code, t)
    var respJson map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&respJson)
    bookId := int(respJson["ID"].(float64) + 0.5)
    resp = getBook(bookId)
    checkResponseCode(expectedCode, resp.Code, t)
    resp = deleteBook(bookId)
    checkResponseCode(expectedCode, resp.Code, t)
    expectedCode = 404
    resp = getBook(bookId)
    checkResponseCode(expectedCode, resp.Code, t)
}

func TestDeleteBadEndpoints(t *testing.T) {
    expectedCode := 404
    badDeleteUri := "/books/?id=1"
    resp := genericRequestNoData(badDeleteUri, "DELETE")
    // expect a redirect, because the ID is on the query string
    checkResponseCode(expectedCode, resp.Code, t)
    resp = getBooks()
    var respJson []map[string]interface{}
    _ = json.NewDecoder(resp.Body).Decode(&respJson)
    selectedBook := respJson[0]
    expectedCode = 200
    resp = deleteBook(int(selectedBook["ID"].(float64)))
    checkResponseCode(expectedCode, resp.Code, t)
    expectedCode = 404
    missingDeleteUri := fmt.Sprintf("/books/%d", int(selectedBook["ID"].(float64)))
    resp = genericRequestNoData(missingDeleteUri, "DELETE")
    checkResponseCode(expectedCode, resp.Code, t)
}


// internals

func checkResponseCode(expectedCode int, receivedCode int, t *testing.T) {
    if expectedCode != receivedCode {
        t.Errorf("response code should be %d, was: %d", expectedCode, receivedCode)
    }
}

const charset = "abcdefghijklmnopqrstuvwxyz" +
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(
  rand.NewSource(time.Now().UnixNano()))

func stringWithCharset(length int, charset string) string {
  b := make([]byte, length)
  for i := range b {
    b[i] = charset[seededRand.Intn(len(charset))]
  }
  return string(b)
}

func getRandomString(length int) string {
  return stringWithCharset(length, charset)
}


func genericRequestNoData(uri string, requestType string) *httptest.ResponseRecorder {
    req, _ := http.NewRequest(requestType, uri, nil)
    resp := httptest.NewRecorder()
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    r.ServeHTTP(resp, req)
    return resp
}

func deleteBook(bookId int) *httptest.ResponseRecorder {
    deleteUri := fmt.Sprintf("/books/%d", bookId)
    req, _ := http.NewRequest("DELETE", deleteUri, nil)
    resp := httptest.NewRecorder()
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    r.ServeHTTP(resp, req)
    return resp
}

func updateBook(bookId int, book Book) *httptest.ResponseRecorder {
    updateUri := fmt.Sprintf("/books/%d", bookId)
    body, _ := json.Marshal(book)
    req, _ := http.NewRequest("PUT", updateUri, bytes.NewReader(body))
    resp := httptest.NewRecorder()
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    r.ServeHTTP(resp, req)
    return resp
}

func getBook(bookId int) *httptest.ResponseRecorder {
    uri := fmt.Sprintf("/books/%d", bookId)
    req, _ := http.NewRequest("GET", uri, nil)
    resp := httptest.NewRecorder()
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    r.ServeHTTP(resp, req)
    return resp
}

func getBooks() *httptest.ResponseRecorder {
    uri := "/books"
    req, _ := http.NewRequest("GET", uri, nil)
    resp := httptest.NewRecorder()
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    r.ServeHTTP(resp, req)
    return resp
}

func createBook(book Book) *httptest.ResponseRecorder {
    body, _ := json.Marshal(book)
    req, _ := http.NewRequest("POST", "/books", bytes.NewReader(body))
    resp := httptest.NewRecorder()
    db := GetDatabaseSession()
    defer db.Close()
    r := GetApiEngine(db)
    r.ServeHTTP(resp, req)
    return resp
}
