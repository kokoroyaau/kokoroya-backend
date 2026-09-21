package middleware

import (
	"slices"
	"strconv"
	"strings"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"kokoroya-backend/internal/authcheck"
	"kokoroya-backend/internal/jwtauth"
	"kokoroya-backend/internal/response"
	"kokoroya-backend/internal/session"
)

const RoleOwner = "owner"

var Pages = []string{
	"dashboard", "labour", "food-cost",
	"employee", "clock-in", "schedule", "salary",
}

func Logger(log *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithWriter(log.Writer())
}

func Recovery(log *logrus.Logger) gin.HandlerFunc {
	return gin.RecoveryWithWriter(log.Writer())
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func RequireAuth(jwtManager *jwtauth.Manager, sessionManager *session.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || tokenStr == "" {
			response.AbortErr(c, 401, "missing bearer token")
			return
		}

		userID, role, jti, err := authcheck.Verify(c.Request.Context(), jwtManager, sessionManager, tokenStr)
		if err != nil {
			response.AbortErr(c, 401, err.Error())
			return
		}

		c.Set("userID", userID)
		c.Set("role", role)
		c.Set("jti", jti)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") != role {
			response.AbortErr(c, 403, "forbidden")
			return
		}
		c.Next()
	}
}

type PermissionLookup func(ctx context.Context, userID int64) ([]string, error)

func RequirePermission(page string, lookup PermissionLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetString("role") == RoleOwner {
			c.Next()
			return
		}

		perms, err := lookup(c.Request.Context(), c.GetInt64("userID"))
		if err != nil || !slices.Contains(perms, page) {
			response.AbortErr(c, 403, "forbidden")
			return
		}
		c.Next()
	}
}

type BranchAccessLookup func(ctx context.Context, userID, branchID int64) (bool, error)

func RequireBranchAccess(lookup BranchAccessLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		branchID, err := strconv.ParseInt(c.GetHeader("X-Branch-ID"), 10, 64)
		if err != nil {
			response.AbortErr(c, 400, "missing or invalid X-Branch-ID header")
			return
		}

		if c.GetString("role") == RoleOwner {
			c.Set("branchID", branchID)
			c.Next()
			return
		}

		ok, err := lookup(c.Request.Context(), c.GetInt64("userID"), branchID)
		if err != nil || !ok {
			response.AbortErr(c, 403, "forbidden")
			return
		}

		c.Set("branchID", branchID)
		c.Next()
	}
}
