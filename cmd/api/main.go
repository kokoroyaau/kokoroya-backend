package main

import "kokoroya-backend/internal/router"

func main() {
	a, err := InitializeApp()
	if err != nil {
		panic(err)
	}
	defer a.DB.Close()
	defer a.Redis.Close()

	a.Logger.Infof("%s starting on port %s (env=%s)", a.Config.App.Name, a.Config.App.Port, a.Config.App.Env)

	engine := router.New(a.DB, a.Redis, a.Config, a.Logger)
	if err := engine.Run(":" + a.Config.App.Port); err != nil {
		a.Logger.Fatal(err)
	}
}
