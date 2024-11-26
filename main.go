package main

import (
	"fmt"
	"net/http"
	"strconv"
	"database/sql"
	"encoding/base64"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type folder struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Database    string  `json:"dbase"`
}

type note struct {
    ID              int     `json:"id"`
    Title           string  `json:"title"`
    MarkdownData    string  `json:"data"`
}

type noteBase struct {
    Name            string
    Notes           []note
}

type baseError struct {
    Code            int     `json:"code"`
    Message         string  `json:"message"`
}

type returnError struct {
    Error           baseError   `json:"error"`
}

type returnData struct {
    Data            any         `json:"data"`
}

// DUMMY DATA
var folders = []folder{
    {ID: 1, Name: "Raspberry_Pi", Database: "pi_kb"},
}

var notes map[string][]note

// GENUINE GLOBALS
var db *sql.DB
const databaseFilename string = "notes.db"

func main() {

    // DUMMY DATA
    testText := "## Get the hostname\n\n```\nhostname\n```\n"
    encoded := base64.StdEncoding.EncodeToString([]byte(testText))
    aNote := new(note)
    aNote.ID = 1
    aNote.Title = "Bash"
    aNote.MarkdownData = encoded
    notes = make(map[string][]note)
    notes["pi_kb"] = make([]note, 0)
    notes["pi_kb"] = append(notes["pi_kb"], *aNote)

    if initDatabase(databaseFilename) != nil {
        fmt.Println("[ERROR] Could not access database")
    }

    router := gin.Default()

    // Return list of folders
    router.GET("/folders", getFolders)

    // Return folder info
    router.GET("/folders/:id", getFolderById)
    router.GET("/folders/:id/notes", getNotesByFolderId)

    // Return note info
    router.GET("/folders/:id/notes/:id2", getNoteById)

    // Run the server
    router.Run("localhost:8080")
}

// Provide a list of folders
func getFolders(c *gin.Context) {

    if len(folders) > 0 {
        c.IndentedJSON(http.StatusOK, makeData(folders))
    } else {
        c.IndentedJSON(http.StatusNotFound, makeError(404, "No folders"))
    }
}

func getFolderById(c *gin.Context) {

    // Convert the string ID to an integer
    id, err := strconv.Atoi(c.Param("id"))
    if err == nil {
        if aFolder := folderById(id); aFolder != nil {
            c.IndentedJSON(http.StatusOK, makeData(aFolder))
            return
        }

        // No folder ID match: issue error
        c.IndentedJSON(http.StatusNotFound, makeError(404, fmt.Sprintf("Folder ID %s not found", c.Param("id"))))
    } else {
        // `id` is not an integer, so see if it's a name
        getFolderByName(c)
    }
}

func getFolderByName(c *gin.Context) {

    name := c.Param("id")
    if aFolder := folderByName(name); aFolder != nil {
        c.IndentedJSON(http.StatusOK, makeData(aFolder))
        return
    }

    // No folder name match: issue error
    c.IndentedJSON(http.StatusNotFound, makeError(404, fmt.Sprintf("Folder %s not found", name)))
}

func getNotesByFolderId(c *gin.Context) {

    // Convert the string ID to an integer
    id, err := strconv.Atoi(c.Param("id"))
    if err == nil {
        if someNotes := notesByFolderId(id); someNotes != nil {
            c.IndentedJSON(http.StatusOK, makeData(someNotes))
            return
        }

        // No folder ID match: issue error
        c.IndentedJSON(http.StatusNotFound, makeError(404, fmt.Sprintf("Folder ID %s not found", c.Param("id"))))
    } else {
        // `id` is not an integer, so see if it's a name
        getFolderByName(c)
    }
}

func getNoteById(c *gin.Context) {

    // Convert the string ID to an integer
    id, err := strconv.Atoi(c.Param("id"))
    if err == nil {
        if someNotes := notesByFolderId(id); someNotes != nil {
            id2, err := strconv.Atoi(c.Param("id2"))
            if err == nil {
                for _, aNote := range *someNotes {
                    if aNote.ID == id2 {
                        c.IndentedJSON(http.StatusOK, makeData(aNote))
                        return
                    }
                }

                // No note match: issue error
                c.IndentedJSON(http.StatusNotFound, makeError(404, fmt.Sprintf("Note ID %s not found", c.Param("id2"))))
            }
        }

        // No folder match: issue error
        c.IndentedJSON(http.StatusNotFound, makeError(404, fmt.Sprintf("Folder ID %s not found", c.Param("id"))))
    }
}

func folderById(id int) *folder {

    for _, aFolder := range folders {
        if aFolder.ID == id {
            return &aFolder
        }
    }

    return nil
}

func notesByFolderId(id int) *[]note {

    if aFolder := folderById(id); aFolder != nil {
        dBase := aFolder.Database
        if value, ok := notes[dBase]; ok == true {
            return &value
        }
    }

    return nil
}

func folderByName(name string) *folder {

    for _, aFolder := range folders {
        if aFolder.Name == name {
            return &aFolder
        }
    }

    return nil
}

func makeData(value any) *returnData {

    result := new(returnData)
    result.Data = value
    return result
}

func makeError(code int, message string) *returnError {

    baseError := new(baseError)
    baseError.Code = code
    baseError.Message = message
    anError := new(returnError)
    anError.Error = *baseError
    return anError
}

func initDatabase(path string) error {

    const createDB string = `
      CREATE TABLE IF NOT EXISTS folders (
      id INTEGER NOT NULL PRIMARY KEY,
      name STRING NOT NULL,
      dbase TEXT
    );`

    // Connect to the database
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return err
    }

    // Attempt to create the database if not present
    if _, err := db.Exec(createDB); err != nil {
        return err
    }

    return nil
}
