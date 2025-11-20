package seeds

import (
	"gonews/internal/core/domain/model"
	"gonews/lib/conv"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	bytes, err := conv.HashPassword("password")

	if err != nil {
		log.Fatal().Err(err).Msg("Error creating password hash")
	}

	admin := model.User {
		Name:     "Admin",
		Email:    "admin@example.com",
		Password: string(bytes),
	}

	if err := db.FirstOrCreate(&admin, model.User{Email: "admin@example.com"}).Error; err != nil {
		log.Fatal().Err(err).Msg("Error seeding admin role")
	} else {
		log.Info().Msg("Admin role seeded Successfully")
	}
}