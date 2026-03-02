package seed

import (
	"golang-grpc-enterprise-demo/internal/model"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Run(db *gorm.DB, logger *zap.Logger) {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		logger.Info("seed skipped — users already exist")
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []model.User{
		{Name: "Alice Wang", Email: "alice@example.com", PasswordHash: string(password)},
		{Name: "Bob Zhang", Email: "bob@example.com", PasswordHash: string(password)},
		{Name: "Charlie Li", Email: "charlie@example.com", PasswordHash: string(password)},
	}

	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			logger.Error("seed user failed", zap.Error(err))
		}
	}
	logger.Info("seed completed", zap.Int("users_created", len(users)))
}
