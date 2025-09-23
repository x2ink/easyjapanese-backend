package middleware

import (
	"easyjapanese/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func User() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求前
		token := c.GetHeader("Authorization")
		tokenData, err := utils.DecryptToken(token)
		if err != nil {
			if err.Error() == "token has invalid claims: token is expired" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"msg":  "Authorization expires",
					"code": 4001,
				})
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg":  "Authentication failed",
				"code": 4002,
			})
			c.Abort()
			return
		}
		c.Set("UserId", tokenData.UserId)
		c.Next()
		// 请求后
	}
}
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求前
		token := c.GetHeader("Authorization")
		tokenData, err := utils.DecryptToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "Authentication failed",
			})
			return
		}
		if tokenData.RoleId != 3 {
			c.JSON(http.StatusForbidden, gin.H{
				"msg": "Insufficient permissions",
			})
			return
		}
		c.Set("UserId", tokenData.UserId)
		c.Next()
		// 请求后
	}
}
func Guest() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		tokenData, err := utils.DecryptToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg": "Authentication failed",
			})
			return
		}
		if tokenData.RoleId != 3 && tokenData.RoleId != 4 {
			c.JSON(http.StatusForbidden, gin.H{
				"msg": "Insufficient permissions",
			})
			return
		}
		c.Set("UserId", tokenData.UserId)
		c.Next()
		// 请求后
	}
}
