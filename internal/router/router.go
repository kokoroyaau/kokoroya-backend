package router

import (
	"context"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"kokoroya-backend/config"
	"kokoroya-backend/internal/jwtauth"
	"kokoroya-backend/internal/middleware"
	"kokoroya-backend/internal/modules/branch"
	"kokoroya-backend/internal/modules/clock"
	"kokoroya-backend/internal/modules/foodcost"
	"kokoroya-backend/internal/modules/labour"
	"kokoroya-backend/internal/modules/schedule"
	"kokoroya-backend/internal/modules/user"
	"kokoroya-backend/internal/session"
)

func New(db *sql.DB, rdb *redis.Client, cfg *config.Config, log *logrus.Logger) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware.Logger(log), middleware.Recovery(log), middleware.CORS())
	api := engine.Group("/v1")

	userRepo := user.NewRepository(db)
	branchRepo := branch.NewRepository(db)
	jwtManager := jwtauth.NewManager(cfg.JWT.Secret, time.Duration(cfg.JWT.AccessTTLMin)*time.Minute)
	sessionManager := session.NewManager(rdb)
	userService := user.NewService(userRepo, jwtManager, sessionManager, log)
	userController := user.NewController(userService)
	branchService := branch.NewService(branchRepo)
	branchController := branch.NewController(branchService)
	labourRepo := labour.NewRepository(db)
	labourService := labour.NewService(labourRepo, branchRepo)
	labourController := labour.NewController(labourService)
	clockRepo := clock.NewRepository(db)
	clockService := clock.NewService(clockRepo, userRepo, labourRepo)
	clockController := clock.NewController(clockService)
	foodCostRepo := foodcost.NewRepository(db)
	foodCostService := foodcost.NewService(foodCostRepo)
	foodCostController := foodcost.NewController(foodCostService)
	scheduleRepo := schedule.NewRepository(db)
	scheduleService := schedule.NewService(scheduleRepo, branchRepo)
	scheduleController := schedule.NewController(scheduleService)
	authMW := middleware.RequireAuth(jwtManager, sessionManager)
	requireBranch := middleware.RequireBranchAccess(branchAccessLookup(branchRepo))
	requireFoodCost := middleware.RequirePermission("food-cost", permissionLookup(userRepo))
	requireLabour := middleware.RequirePermission("labour", permissionLookup(userRepo))
	requireSchedule := middleware.RequirePermission("schedule", permissionLookup(userRepo))
	requireEmployee := middleware.RequirePermission("employee", permissionLookup(userRepo))
	requireSalary := middleware.RequirePermission("salary", permissionLookup(userRepo))
	user.RegisterRoutes(api, userController, authMW, requireEmployee)
	branch.RegisterRoutes(api, branchController, authMW, requireEmployee)
	clock.RegisterRoutes(api, clockController, authMW, requireBranch)
	foodcost.RegisterRoutes(api, foodCostController, authMW, requireBranch, requireFoodCost)
	labour.RegisterRoutes(api, labourController, authMW, requireBranch, requireLabour)
	labour.RegisterSalaryRoutes(api, labourController, authMW, requireBranch, requireSalary)
	schedule.RegisterRoutes(api, scheduleController, authMW, requireBranch, requireSchedule)

	return engine
}

func permissionLookup(repo user.Repository) middleware.PermissionLookup {
	return func(ctx context.Context, userID int64) ([]string, error) {
		u, err := repo.FindBy(ctx, user.Filter{ID: &userID})
		if err != nil {
			return nil, err
		}
		return u.Permissions, nil
	}
}

func branchAccessLookup(repo branch.Repository) middleware.BranchAccessLookup {
	return func(ctx context.Context, userID, branchID int64) (bool, error) {
		return repo.HasAccess(ctx, userID, branchID)
	}
}
