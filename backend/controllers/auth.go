package controllers

import (
	"encoding/json"
	"errors"
	"linkdock/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	web.Controller
}

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type authCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	SecurityQuestion string `json:"securityQuestion"`
	SecurityAnswer   string `json:"securityAnswer"`
}

type recoveryQuestionRequest struct {
	Username string `json:"username"`
}

type resetPasswordRequest struct {
	Username       string `json:"username"`
	SecurityAnswer string `json:"securityAnswer"`
	NewPassword    string `json:"newPassword"`
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// GetUserIdHelper 从 Authorization Header 获取当前用户 ID
func GetUserIdHelper(c *web.Controller) int64 {
	authHeader := c.Ctx.Input.Header("Authorization")
	var idStr string
	if strings.HasPrefix(authHeader, "Bearer ") {
		idStr = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		idStr = c.Ctx.Input.Query("user_id")
	}

	if idStr == "" {
		return 0
	}

	id, _ := strconv.ParseInt(idStr, 10, 64)
	return id
}

func GetCurrentUserHelper(c *web.Controller) (*models.User, bool) {
	userId := GetUserIdHelper(c)
	if userId == 0 {
		return nil, false
	}

	o := orm.NewOrm()
	user := models.User{Id: userId}
	if err := o.Read(&user); err != nil {
		return nil, false
	}
	return &user, true
}

func IsSuperAdminHelper(c *web.Controller) bool {
	user, ok := GetCurrentUserHelper(c)
	return ok && user.IsSuperAdmin
}

func normalizeUsername(username string) string {
	return strings.TrimSpace(strings.ToLower(username))
}

func validateUsername(username string) error {
	if username == "" {
		return errors.New("用户名不能为空")
	}
	if !usernamePattern.MatchString(username) {
		return errors.New("用户名需为 3-20 位字母、数字或下划线")
	}
	if username == "admin" {
		return errors.New("该用户名已被占用")
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码至少需要 8 位")
	}
	hasLetter := false
	hasDigit := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			hasLetter = true
		}
	}
	if !hasLetter || !hasDigit {
		return errors.New("密码需同时包含字母和数字")
	}
	return nil
}

func hashSecret(secret string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func verifySecret(stored string, raw string) bool {
	if stored == "" {
		return false
	}
	if strings.HasPrefix(stored, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(stored), []byte(raw)) == nil
	}
	return stored == raw
}

func getUserByUsername(o orm.Ormer, username string) (*models.User, error) {
	var user models.User
	err := o.QueryTable(new(models.User)).Filter("username", username).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *AuthController) Login() {
	var req authCredentials
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	req.Username = normalizeUsername(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	o := orm.NewOrm()
	user, err := getUserByUsername(o, req.Username)

	// 自动创建第一个管理员
	if err == orm.ErrNoRows && req.Username == "admin" && req.Password == "123456" {
		hashedPassword, hashErr := hashSecret(req.Password)
		if hashErr != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = map[string]interface{}{"success": false, "message": "初始化管理员失败"}
			c.ServeJSON()
			return
		}

		adminUser := models.User{
			Username:     "admin",
			Password:     hashedPassword,
			IsSuperAdmin: true,
			CreatedAt:    time.Now(),
		}
		id, insertErr := o.Insert(&adminUser)
		if insertErr != nil {
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = map[string]interface{}{"success": false, "message": "初始化管理员失败"}
			c.ServeJSON()
			return
		}

		c.Data["json"] = map[string]interface{}{"success": true, "username": "admin", "id": strconv.FormatInt(id, 10), "isSuperAdmin": true}
		c.ServeJSON()
		return
	}

	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "用户名或密码错误"}
		c.ServeJSON()
		return
	}

	if !verifySecret(user.Password, req.Password) {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "用户名或密码错误"}
		c.ServeJSON()
		return
	}

	// 兼容旧的明文密码数据，登录后自动升级为哈希
	if !strings.HasPrefix(user.Password, "$2") {
		if hashedPassword, hashErr := hashSecret(req.Password); hashErr == nil {
			user.Password = hashedPassword
			o.Update(user, "Password")
		}
	}

	c.Data["json"] = map[string]interface{}{
		"success":      true,
		"username":     user.Username,
		"id":           strconv.FormatInt(user.Id, 10),
		"isSuperAdmin": user.IsSuperAdmin,
	}
	c.ServeJSON()
}

func (c *AuthController) Logout() {
	c.Data["json"] = map[string]interface{}{"success": true}
	c.ServeJSON()
}

func (c *AuthController) Me() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Data["json"] = map[string]interface{}{"authenticated": false}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	user := models.User{Id: userId}
	if err := o.Read(&user); err != nil {
		c.Data["json"] = map[string]interface{}{"authenticated": false}
	} else {
		c.Data["json"] = map[string]interface{}{
			"authenticated": true,
			"id":            strconv.FormatInt(user.Id, 10),
			"username":      user.Username,
			"isSuperAdmin":  user.IsSuperAdmin,
		}
	}
	c.ServeJSON()
}

func (c *AuthController) CheckUsername() {
	var req recoveryQuestionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	username := normalizeUsername(req.Username)
	if err := validateUsername(username); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "available": false, "message": err.Error()}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	_, err := getUserByUsername(o, username)
	if err == orm.ErrNoRows {
		c.Data["json"] = map[string]interface{}{"success": true, "available": true, "message": "用户名可使用"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{"success": true, "available": false, "message": "该用户名已被占用"}
	c.ServeJSON()
}

// Register 注册接口
func (c *AuthController) Register() {
	var req registerRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	req.Username = normalizeUsername(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.SecurityQuestion = strings.TrimSpace(req.SecurityQuestion)
	req.SecurityAnswer = strings.TrimSpace(req.SecurityAnswer)

	if err := validateUsername(req.Username); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": err.Error()}
		c.ServeJSON()
		return
	}
	if err := validatePassword(req.Password); err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": err.Error()}
		c.ServeJSON()
		return
	}
	if req.SecurityQuestion == "" {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请选择或填写安全问题"}
		c.ServeJSON()
		return
	}
	if len([]rune(req.SecurityAnswer)) < 2 {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "安全问题答案至少需要 2 个字符"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	if _, err := getUserByUsername(o, req.Username); err == nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "用户名已存在"}
		c.ServeJSON()
		return
	}

	hashedPassword, err := hashSecret(req.Password)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "注册失败，请稍后再试"}
		c.ServeJSON()
		return
	}

	hashedAnswer, err := hashSecret(strings.ToLower(req.SecurityAnswer))
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "注册失败，请稍后再试"}
		c.ServeJSON()
		return
	}

	newUser := models.User{
		Username:         req.Username,
		Password:         hashedPassword,
		SecurityQuestion: req.SecurityQuestion,
		SecurityAnswer:   hashedAnswer,
		IsSuperAdmin:     false,
		CreatedAt:        time.Now(),
	}

	id, err := o.Insert(&newUser)
	if err != nil {
		c.Data["json"] = map[string]interface{}{"success": false, "message": "注册失败，请稍后再试"}
		c.ServeJSON()
		return
	}
	newUser.Id = id

	seedDefaultData(newUser.Id)

	c.Data["json"] = map[string]interface{}{
		"success":  true,
		"message":  "注册成功",
		"id":       strconv.FormatInt(newUser.Id, 10),
		"username": newUser.Username,
	}
	c.ServeJSON()
}

func (c *AuthController) GetRecoveryQuestion() {
	var req recoveryQuestionRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	username := normalizeUsername(req.Username)
	if username == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请输入用户名"}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	user, err := getUserByUsername(o, username)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未找到该用户"}
		c.ServeJSON()
		return
	}

	if strings.TrimSpace(user.SecurityQuestion) == "" || strings.TrimSpace(user.SecurityAnswer) == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "该账户未设置安全找回信息"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{
		"success":          true,
		"securityQuestion": user.SecurityQuestion,
	}
	c.ServeJSON()
}

func (c *AuthController) ResetPassword() {
	var req resetPasswordRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	req.Username = normalizeUsername(req.Username)
	req.SecurityAnswer = strings.TrimSpace(req.SecurityAnswer)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.Username == "" || req.SecurityAnswer == "" || req.NewPassword == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请完整填写重置信息"}
		c.ServeJSON()
		return
	}
	if err := validatePassword(req.NewPassword); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": err.Error()}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	user, err := getUserByUsername(o, req.Username)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusNotFound)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "未找到该用户"}
		c.ServeJSON()
		return
	}

	if !verifySecret(user.SecurityAnswer, strings.ToLower(req.SecurityAnswer)) {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "安全问题答案不正确"}
		c.ServeJSON()
		return
	}

	hashedPassword, err := hashSecret(req.NewPassword)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "重置失败，请稍后再试"}
		c.ServeJSON()
		return
	}

	user.Password = hashedPassword
	if _, err := o.Update(user, "Password"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "重置失败，请稍后再试"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{"success": true, "message": "密码已重置，请重新登录"}
	c.ServeJSON()
}

func (c *AuthController) ChangePassword() {
	userId := GetUserIdHelper(&c.Controller)
	if userId == 0 {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请先登录"}
		c.ServeJSON()
		return
	}

	var req changePasswordRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请求格式错误"}
		c.ServeJSON()
		return
	}

	req.OldPassword = strings.TrimSpace(req.OldPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if req.OldPassword == "" || req.NewPassword == "" {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "请完整填写密码信息"}
		c.ServeJSON()
		return
	}
	if err := validatePassword(req.NewPassword); err != nil {
		c.Ctx.Output.SetStatus(http.StatusBadRequest)
		c.Data["json"] = map[string]interface{}{"success": false, "message": err.Error()}
		c.ServeJSON()
		return
	}

	o := orm.NewOrm()
	user := models.User{Id: userId}
	if err := o.Read(&user); err != nil {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "用户不存在"}
		c.ServeJSON()
		return
	}

	if !verifySecret(user.Password, req.OldPassword) {
		c.Ctx.Output.SetStatus(http.StatusUnauthorized)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "旧密码不正确"}
		c.ServeJSON()
		return
	}

	hashedPassword, err := hashSecret(req.NewPassword)
	if err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "修改失败，请稍后再试"}
		c.ServeJSON()
		return
	}

	user.Password = hashedPassword
	if _, err := o.Update(&user, "Password"); err != nil {
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.Data["json"] = map[string]interface{}{"success": false, "message": "修改失败，请稍后再试"}
		c.ServeJSON()
		return
	}

	c.Data["json"] = map[string]interface{}{"success": true, "message": "密码修改成功"}
	c.ServeJSON()
}
