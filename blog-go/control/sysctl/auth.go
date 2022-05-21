package sysctl

import (
	"blog/conf"
	"blog/internal/rate"
	"blog/internal/token"
	"blog/internal/vcode"
	"blog/model"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/zxysilent/utils"
)

// 防止暴力破解,每秒20次登录限制
var loginLimiter = rate.NewLimiter(20, 5)

const maxErrLogin = 5

// UserLogin doc
// @Tags auth-登陆认证
// @Summary 登陆
// @Accept mpfd
// @Param num formData string true "账号" default(zxysilent)
// @Param pass formData string true "密码" default(zxyslt)
// @Success 200 {object} model.Reply{data=string} "返回数据"
// @Router /api/auth/login [post]
func AuthLogin(ctx echo.Context) error {
	ct, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := loginLimiter.Wait(ct); err != nil {
		return ctx.JSON(utils.Fail("当前登录人数过多,请等待", err.Error()))
	}
	ipt := struct {
		Num    string `json:"num" form:"num"`
		Vcode  string `form:"vcode" json:"vcode"`
		Vreal  string `form:"vreal" json:"vreal"`
		Passwd string `json:"passwd" form:"passwd"`
	}{}
	err := ctx.Bind(&ipt)
	if err != nil {
		return ctx.JSON(utils.ErrIpt("请输入用户名和密码", err.Error()))
	}
	if ipt.Vreal != hmc(ipt.Vcode, "v.c.o.d.e") {
		return ctx.JSON(utils.ErrIpt("请输入正确的验证码"))
	}
	if ipt.Num == "" && len(ipt.Num) > 18 {
		return ctx.JSON(utils.ErrIpt("账号或密码输入错误"))
	}
	mod, has := model.UserLogin(ipt.Num)
	if !has {
		return ctx.JSON(utils.ErrOpt("账号或密码输入错误"))
	}
	now := time.Now()
	// 禁止登陆证 5 分钟
	if mod.Ecount == -1 {
		// 登录时间差
		span := maxErrLogin - int(now.Sub(mod.Ltime).Minutes())
		if span >= 1 { //「」
			return ctx.JSON(utils.Fail("请「" + strconv.Itoa(span) + "」分钟后登录"))
		}
		mod.Ecount = 0
	}
	if mod.Passwd != ipt.Passwd {
		mod.Ltime = now
		mod.Ecount++
		// 错误次数大于 5 锁定
		if mod.Ecount >= maxErrLogin {
			mod.Ecount = -1
			model.UserEdit(mod, "Ltime", "Ecount")
			return ctx.JSON(utils.Fail("登录锁定请「5」分钟后登录"))
		}
		// 小于5 提示剩余次数
		model.UserEdit(mod, "Ltime", "Ecount")
		return ctx.JSON(utils.Fail("密码错误,剩于登录次数：" + strconv.Itoa(maxErrLogin-mod.Ecount)))
	}
	auth := token.Auth{
		Id:     mod.Id,
		RoleId: 0,
		ExpAt:  time.Now().Add(time.Hour * time.Duration(conf.App.TokenExp)).Unix(),
	}
	mod.Ltime = now
	// mod.Ip = ctx.RealIP()
	// model.UserEdit(mod, "Ltime", "Ip", "Ecount")
	return ctx.JSON(utils.Succ("登陆成功", auth.Encode(conf.App.TokenSecret)))
}

// AuthGet doc
// @Tags auth-登陆认证
// @Summary 获取登录信息
// @Param token query string true "凭证"
// @Success 200 {object} model.Reply{data=model.User} "返回数据"
// @Router /adm/auth/get [get]
func AuthGet(ctx echo.Context) error {
	mod, _ := model.UserGet(ctx.Get("uid").(int))
	return ctx.JSON(utils.Succ("auth", mod))
}

// UserLogout doc
// @Tags auth-登陆认证
// @Summary 注销登录
// @Router /api/auth/logout [post]
func UserLogout(ctx echo.Context) error {
	return ctx.HTML(200, "ok")
}

// AuthVcode doc
// @Tags auth-登陆认证
// @Summary 验证码
// @Accept mpfd
// @Success 200 {object} model.Reply{data=string} "返回数据"
// @Router /api/auth/vcode [get]
func AuthVcode(ctx echo.Context) error {
	rnd := utils.RandDigitStr(4)
	out := struct {
		Vcode string `json:"vcode"`
		Vreal string `json:"vreal"`
	}{
		Vcode: vcode.NewImage(rnd).Base64(),
		Vreal: hmc(rnd, "v.c.o.d.e"),
	}
	return ctx.JSON(utils.Succ("succ", out))
}

func hmc(raw, key string) string {
	hm := hmac.New(sha1.New, []byte(key))
	hm.Write([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(hm.Sum(nil))
}

// AuthEdit doc
// @Tags auth-登陆认证
// @Summary 修改个人信息
// @Param name formData string true "名称"
// @Param phone formData string true "号码"
// @Param email formData string true "邮箱"
// @Success 200 {object} model.Reply{data=string} "返回数据"
// @Router /adm/auth/edit [post]
func AuthEdit(ctx echo.Context) error {
	ipt := &model.User{}
	err := ctx.Bind(&ipt)
	if err != nil {
		return ctx.JSON(utils.Fail("输入数据有误", err.Error()))
	}
	ipt.Id = ctx.Get("uid").(int)
	if err := model.UserEdit(ipt, "name", "email", "phone"); err != nil {
		return ctx.JSON(utils.Fail("修改失败", err.Error()))
	}
	return ctx.JSON(utils.Succ("succ"))
}

// AuthPasswd doc
// @Tags auth-登陆认证
// @Summary 修改自己的密码
// @Param opasswd formData string true "旧密码"
// @Param npasswd formData string true "新密码"
// @Success 200 {object} model.Reply{data=string} "返回数据"
// @Router /adm/auth/passwd [post]
func AuthPasswd(ctx echo.Context) error {
	ipt := &struct {
		Opasswd string `form:"opasswd" json:"opasswd"`
		Npasswd string `form:"npassws" json:"npasswd"`
	}{}
	err := ctx.Bind(ipt)
	if err != nil {
		return ctx.JSON(utils.Fail("输入数据有误", err.Error()))
	}
	mod, has := model.UserGet(ctx.Get("uid").(int))
	if !has {
		return ctx.JSON(utils.Fail("输入数据有误,请重试"))
	}
	if mod.Passwd != ipt.Opasswd {
		return ctx.JSON(utils.Fail("原始密码输入错误,请重试"))
	}
	mod.Passwd = ipt.Npasswd
	if err := model.UserEdit(mod, "passwd"); err != nil {
		return ctx.JSON(utils.Fail("密码修改失败", err.Error()))
	}
	return ctx.JSON(utils.Succ("succ"))
}
