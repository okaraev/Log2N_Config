package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func ValidateConfig(config bson.M) error {
	value, ok := config["Team"]
	if !ok || value == "" {
		return fmt.Errorf("cannot find property team")
	}
	value, ok = config["Name"]
	if !ok || value == "" {
		return fmt.Errorf("cannot find property name")
	}
	return nil
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
	configM["Team"] = team
	err = ValidateConfig(configM)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprintf("Error: %s", err)})
		c.Abort()
		return
	}
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
	updatedConfig, err := GetSingleTeamConfig(fmt.Sprint(configM["Team"]), fmt.Sprint(configM["Name"]), ConfigDBConf)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprint(err)})
		c.Abort()
		return
	}
	updatedBytes, err := json.Marshal(updatedConfig)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprint(err)})
		c.Abort()
		return
	}
	err = json.Unmarshal(updatedBytes, &configM)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprint(err)})
		c.Abort()
		return
	}
	configM["UpdateType"] = "Update"
	configM["UpdateTime"] = time.Now()
	err = SendMessage(GlobalConfig.QConnectionString, GlobalConfig.QName, configM)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprint(err)})
		c.Abort()
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
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
