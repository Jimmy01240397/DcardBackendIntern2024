package config

import (
    "os"
    "strconv"
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
    loadenv()
    var err error
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
