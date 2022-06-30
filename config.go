package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var GlobalConfig webconfig
var UserDBConf DBConfig
var ConfigDBConf DBConfig

type httpresponse struct {
	Status  bool
	Message string
}

type TeamConfig struct {
	Name                  string   `bson:"Name" json:"Name"`
	Team                  string   `bson:"Team" json:"Team"`
	LogPattern            string   `bson:"LogPattern" json:"LogPattern"`
	LogSeverity           string   `bson:"LogSeverity" json:"LogSeverity"`
	NotificationMethod    string   `bson:"NotificationMethod" json:"NotificationMethod"`
	LogLogic              string   `bson:"LogLogic" json:"LogLogic"`
	NotificationRecipient []string `bson:"NotificationRecipient" json:"NotificationRecipient"`
	HoldTime              int      `bson:"HoldTime" json:"HoldTime"`
	RetryCount            int      `bson:"RetryCount" json:"RetryCount"`
}

type webconfig struct {
	DBConf            []DBConfig
	QConnectionString string
	QName             string
}

func GetHash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func throw(err error) {
	if err != nil {
		panic(err)
	}
}

func getEnvs() error {
	configDB := os.Getenv("configdb")
	configCol := os.Getenv("ConfigCol")
	ConfigDBCS := os.Getenv("configDBCS")
	userDBCS := os.Getenv("userDBCS")
	userDB := os.Getenv("userdb")
	userCol := os.Getenv("userCol")
	QCS := os.Getenv("QCS")
	QName := os.Getenv("QName")
	if configDB == "" {
		return fmt.Errorf("cannot get environment variable configdb")
	}
	if configCol == "" {
		return fmt.Errorf("cannot get environment variable configcol")
	}
	if ConfigDBCS == "" {
		return fmt.Errorf("cannot get environment variable configdbcs")
	}
	if userDBCS == "" {
		return fmt.Errorf("cannot get environment variable userdbcs")
	}
	if userDB == "" {
		return fmt.Errorf("cannot get environment variable userdb")
	}
	if userCol == "" {
		return fmt.Errorf("cannot get environment variable usercol")
	}
	if QCS == "" {
		return fmt.Errorf("cannot get environment variable qcs")
	}
	if QName == "" {
		return fmt.Errorf("cannot get environment variable qname")
	}
	configdbcsbytes, err := os.ReadFile(ConfigDBCS)
	if err != nil {
		return err
	}
	userdbcsbytes, err := os.ReadFile(userDBCS)
	if err != nil {
		return err
	}
	configConnectionString := strings.Split(string(configdbcsbytes), "\n")[0]
	userConnectionString := strings.Split(string(userdbcsbytes), "\n")[0]
	db1 := DBConfig{Database: configDB, Collection: configCol, Connectionstring: configConnectionString}
	db2 := DBConfig{Database: userDB, Collection: userCol, Connectionstring: userConnectionString}
	dbconf := []DBConfig{db1, db2}
	qcsbytes, err := os.ReadFile(QCS)
	if err != nil {
		return err
	}
	QConnectionString := strings.Split(string(qcsbytes), "\n")[0]
	GlobalConfig.DBConf = dbconf
	GlobalConfig.QConnectionString = QConnectionString
	GlobalConfig.QName = QName
	return nil
}

func main() {
	err := getEnvs()
	throw(err)
	UserDBConf = DBConfig{
		Database:         GlobalConfig.DBConf[1].Database,
		Collection:       GlobalConfig.DBConf[1].Collection,
		Connectionstring: GlobalConfig.DBConf[1].Connectionstring,
	}
	ConfigDBConf = DBConfig{
		Database:         GlobalConfig.DBConf[0].Database,
		Collection:       GlobalConfig.DBConf[0].Collection,
		Connectionstring: GlobalConfig.DBConf[0].Connectionstring,
	}
	for _, dbconf := range GlobalConfig.DBConf {
		err := ValidateDBConfig(dbconf)
		throw(err)
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/1/config", Authenticate, GetmyConfig)
	router.POST("/api/1/config", Authenticate, AddmyConfig)
	router.PUT("/api/1/config", Authenticate, SetmyConfig)
	router.DELETE("/api/1/config", Authenticate, RemovemyConfig)
	router.POST("/api/1/user", Authenticate, SystemAuthorize, AddApiUser)
	router.PUT("/api/1/user", Authenticate, SystemAuthorize, SetApiUser)
	router.DELETE("/api/1/user", Authenticate, SystemAuthorize, RemoveApiUser)
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		throw(fmt.Errorf("cannot find http_port environment variable"))
	}
	router.Run(":" + port)
}
