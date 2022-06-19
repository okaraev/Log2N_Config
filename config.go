package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
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

func SendMessage(amqpServerURL string, QName string, i interface{}) error {
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		return err
	}
	defer connectRabbitMQ.Close()
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channelRabbitMQ.Close()
	bytes, err := json.Marshal(i)
	if err != nil {
		return err
	}
	message := amqp.Publishing{
		ContentType:  "application/json",
		Body:         bytes,
		DeliveryMode: 2,
	}
	err = channelRabbitMQ.Publish("", QName, false, false, message)
	if err != nil {
		return err
	}
	return nil
}

func GetHash(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func Authenticate(c *gin.Context) {
	user, password, ok := c.Request.BasicAuth()
	if !ok {
		c.Abort()
		c.IndentedJSON(403, httpresponse{Status: false, Message: "Not authecticated"})
		return
	}
	userAccount, err := GetTeamUser(user, UserDBConf)
	if err != nil || userAccount.Password != GetHash(password) {
		fmt.Printf("User: %s, Account Password: %s\n", userAccount.Name, userAccount.Password)
		if err != nil {
			log.Println(err)
		}
		c.Abort()
		c.IndentedJSON(403, httpresponse{Status: false, Message: "Not authecticated"})
	} else {
		c.Params = append(c.Params, gin.Param{Key: "Team", Value: userAccount.Team})
	}
}

func PasswordComplexityCheck(password string) bool {
	regx, err := regexp.MatchString("[a-z]", password)
	if err != nil || !regx {
		return false
	}
	regx, err = regexp.MatchString("[A-Z]", password)
	if err != nil || !regx {
		return false
	}
	regx, err = regexp.MatchString("\\d", password)
	if err != nil || !regx {
		return false
	}
	if len(password) < 8 {
		return false
	}
	return true
}

func SystemAuthorize(c *gin.Context) {
	user, password, _ := c.Request.BasicAuth()
	userAccount, err := GetTeamUser(user, UserDBConf)
	if err != nil || userAccount.Team != "System" || userAccount.Password != GetHash(password) {
		if err != nil {
			log.Println(err)
		}
		c.Params = append(c.Params, gin.Param{Key: "isAuthorized", Value: "false"})
		return
	}
}

func GetmyConfig(c *gin.Context) {
	team := c.Params.ByName("Team")
	configs, err := GetTeamConfig(team, ConfigDBConf)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprint(err)})
		return
	}
	c.IndentedJSON(200, configs)
}

func GetAllConfigs(c *gin.Context) {
	isSystemAuthorized := c.Params.ByName("isAuthorized")
	if isSystemAuthorized == "false" {
		c.IndentedJSON(403, httpresponse{Status: false, Message: "Not authorized"})
		return
	}
	configs, err := GetTeamConfig("%", ConfigDBConf)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprint(err)})
		return
	}
	c.IndentedJSON(200, configs)
}

func GetConfigbyTeam(c *gin.Context) {
	team := c.Param("Team")
	teamconfigs, err := GetTeamConfig(team, ConfigDBConf)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": err})
	}
	c.IndentedJSON(http.StatusOK, teamconfigs)
}

func AddmyConfig(c *gin.Context) {
	team := c.Params.ByName("Team")
	configM := bson.M{}
	err := c.BindJSON(&configM)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "invalid character") || strings.Contains(message, "cannot unmarshal") {
			message = "post body must be in json format"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s,Method: %s, Stage: BindJSON, func: AddmyConfig, Message: %s", apiuser, c.Request.Method, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	bytes, err := json.Marshal(configM)
	if err != nil {
		panic(err)
	}
	config := TeamConfig{}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		panic(err)
	}
	config.Team = team
	if config.LogSeverity == "" || config.LogLogic == "" || config.Name == "" || config.NotificationMethod == "" || config.NotificationRecipient == nil || len(config.NotificationRecipient) == 0 {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "LogSeverity, LogLogic, Name, NotificationMethod, NotificationRecipient fields cannot be null"})
		c.Abort()
		return
	}
	err = AddTeamConfig(config, ConfigDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "dup key") {
			message = "Given Configuration Name already exist"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: AddTeamConfig, func: AddmyConfig, Message: %s", apiuser, c.Request.Method, configM, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
	configM["Team"] = team
	configM["UpdateType"] = "Add"
	configM["UpdateTime"] = time.Now()
	err = SendMessage(GlobalConfig.QConnectionString, GlobalConfig.QName, configM)
	if err != nil {
		log.Println(err)
	}
}

func SetmyConfig(c *gin.Context) {
	team := c.Params.ByName("Team")
	configM := bson.M{}
	err := c.BindJSON(&configM)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "invalid character") || strings.Contains(message, "cannot unmarshal") {
			message = "post body must be in json format"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Stage: BindJSON, func: SetmyConfig, Message: %s", apiuser, c.Request.Method, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	structure := TeamConfig{}
	ref := reflect.TypeOf(structure)
	refM := reflect.ValueOf(configM).MapKeys()
	for in := 0; in < len(refM); in++ {
		valid := false
		for i := 0; i < ref.NumField(); i++ {
			if refM[in].String() == ref.Field(i).Name {
				valid = true
				break
			}
		}
		if !valid {
			message := fmt.Sprintf("Cannot validate field: %s", refM[in].String())
			c.IndentedJSON(424, httpresponse{Status: false, Message: message})
			c.Abort()
			return
		}
	}
	configM["Team"] = team
	err = SetTeamConfig(configM, ConfigDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if message == "no document found to update" {
			message = fmt.Sprintf("no configuration found with name: %s", configM["Name"].(string))
		} else if message == "nothing to update" {
			message = "no difference between given and stored configuration"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: SetTeamConfig, func: SetmyConfig, Message: %s", apiuser, c.Request.Method, configM, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
	configM["UpdateType"] = "Update"
	configM["UpdateTime"] = time.Now()
	err = SendMessage(GlobalConfig.QConnectionString, GlobalConfig.QName, configM)
	if err != nil {
		log.Println(err)
	}
}

func RemovemyConfig(c *gin.Context) {
	team := c.Params.ByName("Team")
	configM := bson.M{}
	err := c.BindJSON(&configM)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "invalid character") || strings.Contains(message, "cannot unmarshal") {
			message = "post body must be in json format"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Stage: BindJSON, func: RemovemyConfig, Message: %s", apiuser, c.Request.Method, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	val, ok := configM["Name"]
	if !ok || val == "" {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Name field cannot be null or empty"})
		c.Abort()
		return
	}
	err = RemoveTeamConfig(configM["Name"].(string), ConfigDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if message == "nothing to delete" {
			message = fmt.Sprintf("There is no configuration with name: %s", configM["Name"].(string))
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: RemoveTeamConfig, func: RemovemyConfig, Message: %s", apiuser, c.Request.Method, configM, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
	configM["Team"] = team
	configM["UpdateType"] = "Delete"
	configM["UpdateTime"] = time.Now()
	err = SendMessage(GlobalConfig.QConnectionString, GlobalConfig.QName, configM)
	if err != nil {
		log.Println(err)
	}
}

func AddApiUser(c *gin.Context) {
	isSystemAuthorized := c.Params.ByName("isAuthorized")
	if isSystemAuthorized == "false" {
		c.IndentedJSON(403, httpresponse{Status: false, Message: "Not authorized"})
		return
	}
	user := Account{}
	err := c.BindJSON(&user)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "invalid character") || strings.Contains(message, "cannot unmarshal") {
			message = "post body must be in json format"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Stage: BindJSON, func: AddApiUser, Message: %s", apiuser, c.Request.Method, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	if user.Team == "" || user.Password == "" || user.Name == "" {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Team, Name, Password fields cannot be null"})
		return
	}
	pwdCheck := PasswordComplexityCheck(user.Password)
	if !pwdCheck {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Provided password doesn't meet required complexity. Please see documentation"})
		return
	}
	user.Password = GetHash(user.Password)
	err = AddTeamUser(user, UserDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "Team.Users index: name dup key") {
			message = fmt.Sprintf("There is already have user with name %s", user.Name)
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: AddTeamUser, func: AddApiUser, Message: %s", apiuser, c.Request.Method, user, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
}

func SetApiUser(c *gin.Context) {
	isSystemAuthorized := c.Params.ByName("isAuthorized")
	userName, _, _ := c.Request.BasicAuth()
	user := bson.M{}
	err := c.BindJSON(&user)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "invalid character") || strings.Contains(message, "cannot unmarshal") {
			message = "post body must be in json format"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Stage: BindJSON, func: SetApiUser, Message: %s", apiuser, c.Request.Method, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		return
	}
	if _, ok := user["Name"]; !ok || user["Name"] == "" {
		user["Name"] = userName
	} else if user["Name"] != userName && isSystemAuthorized == "false" {
		c.IndentedJSON(403, httpresponse{Status: false, Message: "You don't have permissions for this change"})
		return
	}
	_, ok := user["Password"]
	pwdCheck := false
	if ok {
		pwdCheck = PasswordComplexityCheck(user["Password"].(string))
	}
	if !pwdCheck {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Provided password doesn't meet required complexity. Please see documentation"})
		return
	}
	user["Password"] = GetHash(user["Password"].(string))
	err = SetTeamUser(user, UserDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if message == "no document found to update" {
			message = fmt.Sprintf("no user found with name: %s", user["Name"].(string))
		} else if message == "nothing to update" {
			message = "same password provided"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: SetTeamUser, func: SetApiUser, Message: %s", apiuser, c.Request.Method, user, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
}

func RemoveApiUser(c *gin.Context) {
	isSystemAuthorized := c.Params.ByName("isAuthorized")
	if isSystemAuthorized == "false" {
		c.IndentedJSON(403, httpresponse{Status: false, Message: "Not authorized"})
		return
	}
	user := bson.M{}
	if err := c.BindJSON(&user); err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "invalid character") || strings.Contains(message, "cannot unmarshal") {
			message = "post body must be in json format"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Stage: BindJSON, func: RemoveApiUser, Message: %s", apiuser, c.Request.Method, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		return
	}
	myUser, ok := user["Name"]
	if !ok || myUser == "" {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Team, Name, Password fields cannot be null"})
		return
	}
	err := RemoveTeamUser(myUser.(string), UserDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if message == "nothing to delete" {
			message = fmt.Sprintf("no user found with name: %s", myUser.(string))
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: RemoveTeamUser, func: RemoveApiUser, Message: %s", apiuser, c.Request.Method, user, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func getEnvs() error {
	configDB := os.Getenv("configdb")
	configCol := os.Getenv("ConfigCol")
	DBCS := os.Getenv("DBCS")
	userDB := os.Getenv("userdb")
	userCol := os.Getenv("userCol")
	QCS := os.Getenv("QCS")
	QName := os.Getenv("QName")
	if configDB == "" {
		return fmt.Errorf("cannot get environment variable teamdb")
	}
	if configCol == "" {
		return fmt.Errorf("cannot get environment variable configcol")
	}
	if DBCS == "" {
		return fmt.Errorf("cannot get environment variable dbcs")
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
	Config, err := os.ReadFile(DBCS)
	checkError(err)
	configConnectionString := strings.Split(string(Config), "\r\n")[0]
	userConnectionString := strings.Split(string(Config), "\n")[1]
	db1 := DBConfig{Database: configDB, Collection: configCol, Connectionstring: configConnectionString}
	db2 := DBConfig{Database: userDB, Collection: userCol, Connectionstring: userConnectionString}
	dbconf := []DBConfig{db1, db2}
	Config, err = os.ReadFile(QCS)
	checkError(err)
	QConnectionString := strings.Split(string(Config), "\n")[0]
	GlobalConfig.DBConf = dbconf
	GlobalConfig.QConnectionString = QConnectionString
	GlobalConfig.QName = QName
	return nil
}

func main() {
	err := getEnvs()
	checkError(err)
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
		checkError(err)
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
	router.Run(":" + port)
}
