package router

import (
	"blog/control/appctl"
	"blog/control/sysctl"

	"github.com/labstack/echo/v4"
)

// admRouter 登录访问
func admRouter(adm *echo.Group) {
	adm.GET("/auth/get", sysctl.AuthGet)             // 登陆信息
	adm.POST("/auth/edit", sysctl.AuthEdit)          // 修改信息
	adm.POST("/auth/passwd", sysctl.AuthPasswd)      // 修改密码
	adm.GET("/status/goinfo", sysctl.StatusGoinfo)   // 环境信息
	adm.GET("/status/appinfo", sysctl.StatusAppinfo) // 统计信息
	adm.POST("/upload/file", sysctl.UploadFile)      // 文件上传
	adm.POST("/upload/image", sysctl.UploadImage)    // 图片上传
	adm.POST("/global/edit", sysctl.GlobalEdit)      // 配置修改
	adm.POST("/dict/add", sysctl.DictAdd)            // 添加字典
	adm.POST("/dict/edit", sysctl.DictEdit)          // 编辑字典
	adm.POST("/dict/drop", sysctl.DictDrop)          // 删除字典
	adm.POST("/cate/add", appctl.CateAdd)            // 添加分类
	adm.POST("/cate/edit", appctl.CateEdit)          // 编辑分类
	adm.POST("/cate/drop", appctl.CateDrop)          // 删除分类
	adm.POST("/tag/add", appctl.TagAdd)              // 添加标签
	adm.POST("/tag/edit", appctl.TagEdit)            // 编辑标签
	adm.POST("/tag/drop", appctl.TagDrop)            // 删除标签
	adm.POST("/post/add", appctl.PostAdd)            // 添加文章
	adm.POST("/post/edit", appctl.PostEdit)          // 编辑文章
	adm.POST("/post/drop", appctl.PostDrop)          // 删除文章
	adm.POST("/page/add", appctl.PageAdd)            // 添加页面
	adm.POST("/page/edit", appctl.PageEdit)          // 编辑页面
	adm.POST("/page/drop", appctl.PageDrop)          // 删除页面
}
