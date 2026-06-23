package dal

import (
	"context"
	"fmt"
	"os"

	"origadmin/application/origstudio/internal/data/entity"
	"origadmin/application/origstudio/internal/data/entity/user"
	"origadmin/application/origstudio/internal/pkg/idutil"

	"github.com/origadmin/runtime/log"
	hash "github.com/origadmin/toolkits/crypto/hash"
)

func SeedAdminUser(ctx context.Context, db *entity.Client) error {
	count, err := db.User.Query().Count(ctx)
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	if count > 0 {
		return nil
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin123"
	}
	if adminEmail == "" {
		adminEmail = "admin@origstudio.local"
	}

	crypto, err := hash.NewCrypto("bcrypt")
	if err != nil {
		return fmt.Errorf("failed to create crypto: %w", err)
	}

	hashedPassword, err := crypto.Hash(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	_, err = db.User.Create().
		SetID(idutil.GenUUID()).
		SetUsername(adminUsername).
		SetName("Administrator").
		SetEmail(adminEmail).
		SetPassword(hashedPassword).
		SetSlug("u-admin").
		SetStatus(user.StatusACTIVE).
		SetRole(user.RoleAdmin).
		SetIsSuperuser(true).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Infof("Created default admin user: %s (password: %s)", adminUsername, adminPassword)
	return nil
}
