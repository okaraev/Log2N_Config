package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func SetConfigValidate(config bson.M) error {
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

func AddConfigValidation(config bson.M) error {
	props := []string{"LogSeverity", "LogLogic", "Name", "NotificationMethod", "NotificationRecipient"}
	for _, prop := range props {
		val, ok := config[prop]
		if val == "" || !ok {
			return fmt.Errorf("logseverity, loglogic, name, notificationmethod, notificationrecipient fields cannot be null or empty")
		}
	}
	return nil
}

func AddUserValidation(user bson.M) error {
	for _, prop := range []string{"Team", "Name", "Password"} {
		val, ok := user[prop]
		if val == "" || !ok {
			return fmt.Errorf("Team, Name, Password fields cannot be null")
		}
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
	filter := bson.M{"Team": team}
	configs, err := GetDocument(filter, ConfigDBConf)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprint(err)})
		return
	}
	c.IndentedJSON(200, configs)
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
	err = AddConfigValidation(configM)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprintln(err)})
		c.Abort()
		return
	}
	configM["Team"] = team

	err = AddDocument(configM, ConfigDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "dup key") {
			message = "Given Configuration Name already exist"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: AddTeamConfig, func: AddDocument, Message: %s", apiuser, c.Request.Method, configM, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	c.IndentedJSON(200, httpresponse{Status: true, Message: ""})
	configM["UpdateType"] = "Add"
	configM["UpdateTime"] = time.Now()
	err = myBreaker.Do(GlobalConfig.QConnectionString, GlobalConfig.QName, configM)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Cannot process request"})
		c.Abort()
		return
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
	err = SetConfigValidate(configM)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprintln(err)})
		c.Abort()
		return
	}
	filter := bson.M{"Team": team, "Name": configM["Name"]}
	update := bson.D{
		{Key: "$set", Value: configM},
	}
	updatedConfig, err := SetGetDocument(filter, update, ConfigDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if message == "no document found to update" {
			message = fmt.Sprintf("no configuration found with name: %s", configM["Name"].(string))
		} else if message == "nothing to update" {
			message = "no difference between given and stored configuration"
		} else {
			apiuser, _, _ := c.Request.BasicAuth()
			errmessage := fmt.Sprintf("Api User: %s, Method: %s, Body: %s, Stage: SetTeamConfig, func: SetGetDocument, Message: %s", apiuser, c.Request.Method, configM, message)
			log.Println(errmessage)
			message = "Unhandled exception. Please contact to Administrator"
		}
		c.IndentedJSON(424, httpresponse{Status: false, Message: message})
		c.Abort()
		return
	}
	updatedConfig["UpdateType"] = "Update"
	updatedConfig["UpdateTime"] = time.Now()
	err = myBreaker.Do(GlobalConfig.QConnectionString, GlobalConfig.QName, updatedConfig)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Cannot process request"})
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
	filter := bson.M{"Name": configM["Name"]}
	err = RemoveDocument(filter, ConfigDBConf)
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
	err = myBreaker.Do(GlobalConfig.QConnectionString, GlobalConfig.QName, configM)
	if err != nil {
		log.Println(err)
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Cannot process request"})
		c.Abort()
		return
	}
}

func AddApiUser(c *gin.Context) {
	isSystemAuthorized := c.Params.ByName("isAuthorized")
	if isSystemAuthorized == "false" {
		c.IndentedJSON(403, httpresponse{Status: false, Message: "Not authorized"})
		return
	}
	user := bson.M{}
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
	err = AddUserValidation(user)
	if err != nil {
		c.IndentedJSON(424, httpresponse{Status: false, Message: fmt.Sprintln(err)})
		return
	}
	pwdCheck := PasswordComplexityCheck(user["Password"].(string))
	if !pwdCheck {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Provided password doesn't meet required complexity. Please see documentation"})
		return
	}
	user["Password"] = GetHash(user["Password"].(string))
	err = AddDocument(user, UserDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if strings.Contains(message, "Team.Users index: name dup key") {
			message = fmt.Sprintf("There is already have user with name %s", user["Name"])
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
	filter := bson.M{"Name": user["Name"]}
	update := bson.D{
		{Key: "$set", Value: bson.M{"Password": user["Password"]}},
	}
	err = SetDocument(filter, update, UserDBConf)
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
	val, ok := user["Name"]
	if !ok || val == "" {
		c.IndentedJSON(424, httpresponse{Status: false, Message: "Name field cannot be null"})
		return
	}
	filter := bson.M{"Name": user["Name"]}
	err := RemoveDocument(filter, UserDBConf)
	if err != nil {
		message := fmt.Sprint(err)
		if message == "nothing to delete" {
			message = fmt.Sprintf("no user found with name: %s", user["Name"].(string))
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
