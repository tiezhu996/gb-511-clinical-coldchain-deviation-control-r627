package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/config"
	"github.com/blueship581/clinical-coldchain-deviation-control/backend/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type Principal struct {
	Subject     string
	Username    string
	DisplayName string
	Role        string
}

var roleRank = map[string]int{
	"viewer": 1, "operator": 2, "reviewer": 3, "admin": 4,
}

func Authenticate(cfg config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(value, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing_token"})
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(value, "Bearer "))
		token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		}, jwt.WithIssuer(cfg.AppName), jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_claims"})
			return
		}
		principal := Principal{
			Subject: stringClaim(claims, "sub"), Username: stringClaim(claims, "username"),
			DisplayName: stringClaim(claims, "name"), Role: stringClaim(claims, "role"),
		}
		userID, idErr := strconv.ParseUint(principal.Subject, 10, 64)
		if idErr != nil || userID == 0 || principal.Username == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "incomplete_claims"})
			return
		}
		var user model.User
		if err := db.WithContext(c.Request.Context()).Where("id = ? AND username = ? AND active = ?", userID, principal.Username, true).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "inactive_principal"})
			return
		}
		var role model.Role
		if err := db.WithContext(c.Request.Context()).Where("code = ? AND active = ?", user.Role, true).First(&role).Error; err != nil || roleRank[role.Code] == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "inactive_role"})
			return
		}
		principal.Username, principal.DisplayName, principal.Role = user.Username, user.DisplayName, role.Code
		c.Set("subject", principal.Subject)
		c.Set("username", principal.Username)
		c.Set("displayName", principal.DisplayName)
		c.Set("role", principal.Role)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(c *gin.Context) {
		role, ok := c.Get("role")
		roleName, valid := role.(string)
		if !ok || !valid || !allowed[roleName] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func RequireMinimumRole(minimum string) gin.HandlerFunc {
	minimumRank := roleRank[minimum]
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleName, _ := role.(string)
		if minimumRank == 0 || roleRank[roleName] < minimumRank {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient_role"})
			return
		}
		c.Next()
	}
}

func CurrentPrincipal(c *gin.Context) (Principal, bool) {
	subject, subjectOK := c.Get("subject")
	username, usernameOK := c.Get("username")
	displayName, _ := c.Get("displayName")
	role, roleOK := c.Get("role")
	principal := Principal{
		Subject: stringFromContext(subject), Username: stringFromContext(username),
		DisplayName: stringFromContext(displayName), Role: stringFromContext(role),
	}
	return principal, subjectOK && usernameOK && roleOK && principal.Username != ""
}

func stringFromContext(value any) string {
	text, _ := value.(string)
	return text
}

func stringClaim(claims jwt.MapClaims, key string) string {
	value, _ := claims[key].(string)
	return value
}
