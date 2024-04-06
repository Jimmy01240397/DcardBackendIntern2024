package config

import (
    "os"
    "log"
    "strconv"
    "github.com/joho/godotenv"
)

var Debug bool
var Port string
var DBservice string
var DBuser string
var DBpasswd string
var DBhost string
var DBport string
var DBname string

func init() {
    err := godotenv.Load()
    if err != nil {
        log.Panicln("Error loading .env file")
    }
    debugstr, exists := os.LookupEnv("DEBUG")
    if !exists {
        Debug = false
    } else {
        Debug, err = strconv.ParseBool(debugstr)
        if err != nil {
            Debug = false
        }
    }
    Port = os.Getenv("PORT")
    DBservice = os.Getenv("DBSERVICE")
    DBuser = os.Getenv("DBUSER")
    DBpasswd = os.Getenv("DBPASSWD")
    DBhost = os.Getenv("DBHOST")
    DBport = os.Getenv("DBPORT")
    DBname = os.Getenv("DBNAME")
}
