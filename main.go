package main
import "fmt"
import "os"
import "time"
import "github.com/jinzhu/gorm"
import _ "github.com/jinzhu/gorm/dialects/postgres"
import "github.com/gin-gonic/gin"

// Book model based on crash course in GORM docs
// http://doc.gorm.io/models.html
type Book struct {
    gorm.Model
    Title       string      `json:"title" gorm:"size:255"`
    Author      string      `json:"author" gorm:"size:255"`
    Publisher   string      `json:"publisher" gorm:"size:255"`
    PublishDate string      `json:"publish_date"`
    Rating      int         `json:"rating"`
    CheckedOut  bool        `json:"checked_out"`
}

// Book model Merge function used in updating
func (b *Book) Merge(b2 Book) {
    if b2.Title != "" {
        b.Title = b2.Title
    }
    if b2.Author != "" {
        b.Author = b2.Author
    }
    if b2.Publisher != "" {
        b.Publisher = b2.Publisher
    }
    if b2.PublishDate != "" {
        b.PublishDate = b2.PublishDate
    }
    if b2.Rating != 0 {
        b.Rating = b2.Rating
    }
    if b2.CheckedOut != b.CheckedOut {
        b.CheckedOut = b2.CheckedOut
    }
}

// Launchs API server on port 8080
func main() {
    db := GetDatabaseSession()
    defer db.Close()
    GetApiEngine(db).Run(":8080")
}

// Returns Gin API engine with middleware and routes defined, also 
// adds database session to middleware for handler functions
func GetApiEngine(db *gorm.DB) *gin.Engine {
    db.AutoMigrate(&Book{})
    r := gin.Default()
    r.Use(databaseMiddleware(db))
    r.GET("/books", GetBooks)
    r.GET("/books/:id", GetBook)
    r.POST("/books", CreateBook)
    r.PUT("/books/:id", UpdateBook)
    r.DELETE("books/:id", DeleteBook)
    return r
}

// Returns gorm.DB database session
func GetDatabaseSession() *gorm.DB {
    db, err := gorm.Open(
        "postgres",
        fmt.Sprintf(
            "host=%s user=foo dbname=gobooks password=bar sslmode=disable",
            os.Getenv("POSTGRES_URL")))
    if err != nil {
        panic("failed to connect to database")
    }
    return db
}

// Gin Middleware function to expose database to API handler functions
func databaseMiddleware(db *gorm.DB) gin.HandlerFunc {
    // adds database session to the gin context passed into our handlers
    return func(c *gin.Context) {
        c.Set("database", db)
        c.Next()
    }
}

/* Creates a single Book
    * note: publish_date is required 

    HTTP POST URI format:
        [api_address]:8080/books/

    Example HTTP POST data structure follows:
        {
            "title": "Foo",
            "author": "Bar",
            "publish_date": "1981-Sep-26",
            "rating": 1,
            "publisher": "Foobar Publishing"
        }

    Returns Book map[string]interface{} JSON

    
*/
func CreateBook(c *gin.Context) {
    var book Book
    db := getDatabaseFromGinContext(c)
    c.BindJSON(&book)
    ok, reason := validateBook(&book)
    if !ok {
        c.JSON(400, reason)
        return
    }
    db.Create(&book)
    c.JSON(200, book)
}

/* Query a single Book
    HTTP GET URI format:
        [api_address]:8080/books/[book_id]

    Returns Book map[string]interface{} JSON
*/
func GetBook(c *gin.Context) {
    var book Book
    db := getDatabaseFromGinContext(c)
    id := c.Params.ByName("id")
    err := db.First(&book, id).Error
    if err != nil {
        c.AbortWithStatus(404)
        return
    } else {
        c.JSON(200, book)
    }
}

/* Query all the Books
    HTTP GET URI format:
        [api_address]:8080/books/

    Returns Book []map[string]interface{} JSON
*/
func GetBooks(c *gin.Context) {
    var books []Book
    db := getDatabaseFromGinContext(c)
    db.Find(&books)
    c.JSON(200, books)
}

/* Updates a single Book
    HTTP PUT URI format:
        [api_address]:8080/books/[book_id]

    Example HTTP PUT data structure follows:
        {
            "checked_out": true
        }
    
    Returns Book map[string]interface{} JSON
*/
func UpdateBook(c *gin.Context) {
    var originalBook Book
    var updatedBook Book
    db := getDatabaseFromGinContext(c)
    id := c.Params.ByName("id")
    err := db.First(&originalBook, id).Error
    if err != nil {
        c.AbortWithStatus(404)
        return
    }
    c.BindJSON(&updatedBook)
    if updatedBook.PublishDate != "" || updatedBook.Rating != 0 {
        ok, reason := validateBook(&updatedBook)
        if !ok {
            c.JSON(400, reason)
            return
        }
    }
    originalBook.Merge(updatedBook)
    db.Save(&originalBook)
    c.JSON(200, originalBook)
}

/* Deletes a single Book
    HTTP DELETE URI format:
        [api_address]:8080/books/[book_id]

    Returns Book map[string]interface{} JSON
*/
func DeleteBook(c *gin.Context) {
    var book Book
    db := getDatabaseFromGinContext(c)
    id := c.Params.ByName("id")
    err := db.Where("ID = ?", id).First(&book).Error
    if err != nil {
        c.AbortWithStatus(404)
        return
    }
    _ = db.Where("id = ?", id).Delete(&book)
    c.JSON(200, book)
}

// internals

// Grab the database from the gin context
// Used to be more DRY in our handlers
func getDatabaseFromGinContext(c *gin.Context) *gorm.DB {
    db, ok := c.MustGet("database").(*gorm.DB)
    if !ok {
        panic("failed to get database")
    }
    return db
}

// run validation
func validateBook(book *Book) (bool, map[string]interface{}) {
    var resultMap = make(map[string]interface{})
    if len(book.Title) > 255 {
        resultMap["title"] =  "too long"
    }
    if len(book.Author) > 255 {
        resultMap["author"] =  "too long"
    }
    if len(book.Publisher) > 255 {
        resultMap["publisher"] = "too long"
    }
    if book.Rating > 3 || book.Rating < 1 && book.Rating != 0 {
        resultMap["rating"] = book.Rating
    }
    _, err := time.Parse("2006-Jan-02", book.PublishDate)
    if err != nil {
        resultMap["publish_date"] = book.PublishDate
    }
    if len(resultMap) == 0 {
        return true, resultMap
    } else {
        return false, resultMap
    }
}
