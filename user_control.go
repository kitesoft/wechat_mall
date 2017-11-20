package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gopkg.in/kataras/iris.v6"
)

func GenWeAppLoginUrl(code string) string {
	var buf bytes.Buffer
	buf.WriteString(srvConfig.Wechat.CodeToSessURL)
	v := url.Values{
		"appid":      {srvConfig.Wechat.AppID},
		"secret":     {srvConfig.Wechat.Secret},
		"js_code":    {code},
		"grant_type": {"authorization_code"},
	}

	if strings.Contains(srvConfig.Wechat.CodeToSessURL, "?") {
		buf.WriteByte('&')
	} else {
		buf.WriteByte('?')
	}
	buf.WriteString(v.Encode())
	return buf.String()
}

type WechatSessionResp struct {
	OpenID     string `json:"openid" valid:"required"`
	SessionKey string `json:"session_key" valid:"required"`
	Unionid    string `json:"unionid"`
}

// WeAppLogin 微信小程序登录
func WeAppLogin(ctx *iris.Context) {

	code := ctx.FormValue("code")
	if code == "" {
		SendErrJSON("code不能为空", ctx)
		return
	}

	resp, err := http.Get(GenWeAppLoginUrl(code))
	if err != nil {
		fmt.Println(err.Error())
		SendErrJSON("error", ctx)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		SendErrJSON("error", ctx)
		return
	}

	data := WechatSessionResp{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Println(err.Error())
		SendErrJSON("error", ctx)
		return
	}

	session := ctx.Session()
	session.Set("weAppOpenID", data.OpenID)
	session.Set("weAppSessionKey", data.SessionKey)

	resData := iris.Map{}
	resData[srvConfig.Server.SessionID] = session.ID()
	ctx.JSON(iris.StatusOK, iris.Map{
		"errNo": 200,
		"msg":   "success",
		"data":  resData,
	})
}

type EncryptedUser struct {
	EncryptedData string `json:"encryptedData"`
	IV            string `json:"iv"`
}

// SetWeAppUserInfo 设置小程序用户加密信息 将账号写入到数据库中
func SetWeAppUserInfo(ctx *iris.Context) {

	var weAppUser EncryptedUser

	if ctx.ReadJSON(&weAppUser) != nil {
		SendErrJSON("参数错误", ctx)
		return
	}
	session := ctx.Session()
	sessionKey := session.GetString("weAppSessionKey")
	if sessionKey == "" {
		SendErrJSON("session error", ctx)
		return
	}

	userInfoStr, err := DecodeWeAppUserInfo(weAppUser.EncryptedData, sessionKey, weAppUser.IV)
	if err != nil {
		fmt.Println(err.Error())
		SendErrJSON("error", ctx)
		return
	}

	var user WeAppUser
	if err := json.Unmarshal([]byte(userInfoStr), &user); err != nil {
		SendErrJSON("error", ctx)
		return
	}

	mallUser := &User{}
	DB.Where("nickname = ?", user.Nickname).First(mallUser)
	if mallUser.ID == 0 {
		//执行插入操作
		mallUser.Nickname = user.Nickname
		mallUser.OpenID = user.OpenID
		DB.Create(&mallUser)
	}

	session.Set("weAppUser", user)
	ctx.JSON(iris.StatusOK, iris.Map{
		"errNo": 2000,
		"msg":   "success",
		"data":  iris.Map{},
	})
	return
}

//更新用户信息
func UpdateUserInfo(ctx *iris.Context) {

	var user User
	var updateMap iris.Map

	if ctx.ReadJSON(&updateMap) != nil {
		SendErrJSON("参数错误", ctx)
		return
	}

	//TODO 需要看一下是否能够正确获取
	session := ctx.Session()
	if session == nil {
		SendErrJSON("seesion 为空", ctx)
		return
	}
	all := session.GetAll()
	fmt.Printf("%v", all)

	openId := session.GetString("weAppOpenID")

	DB.Model(&user).Where("open_id = ?", openId).Updates(updateMap)

	ctx.JSON(iris.StatusOK, iris.Map{
		"errNo": 2000,
		"msg":   "success",
		"data":  iris.Map{},
	})
	return
}

//获取用户信息
func GetUserInfo(ctx *iris.Context) {
	var user User

	session := ctx.Session()
	if session == nil {
		SendErrJSON("seesion 为空", ctx)
		return
	}

	openId := session.GetString("weAppOpenID")

	DB.Model(&user).Where("open_id = ?", openId)

	ctx.JSON(iris.StatusOK, iris.Map{
		"errNo": 2000,
		"msg":   "success",
		"data":  user,
	})
	return
}
