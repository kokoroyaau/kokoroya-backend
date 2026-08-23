// Command seed creates the owner user, idempotently, from OwnerConfig.
// There is no public register endpoint — this is the only way an owner
// account comes into existence.
package main

import (
	"golang.org/x/crypto/bcrypt"

	"kokoroya-backend/config"
	"kokoroya-backend/internal/database"
	"kokoroya-backend/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.New(cfg)

	db, err := database.NewPostgresConnection(cfg, log)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Owner.Password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	res, err := db.Exec(`
		insert into users (name, email, password_hash, role, is_active, permissions)
		values ('Owner', $1, $2, 'owner', true, '{}')
		on conflict (email) do nothing
	`, cfg.Owner.Email, string(hash))
	if err != nil {
		panic(err)
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		log.Infof("owner %s already exists, nothing to do", cfg.Owner.Email)
		return
	}
	log.Infof("seeded owner %s", cfg.Owner.Email)
}
