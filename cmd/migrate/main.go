package main

import (
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ranson-backend/internal/app/ds"
	"ranson-backend/internal/app/dsn"
)

func main() {
	_ = godotenv.Load()

	connString := dsn.FromEnv()
	if connString == "" {
		log.Fatal("не найден .env или в нем нет DB_HOST: запускайте команду из папки Web-backend_medicine_Ranson")
	}

	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		log.Fatalf("не удалось подключиться к БД: %v", err)
	}

	err = db.AutoMigrate(
		&ds.PancreatitisUser{},
		&ds.PancreatitisSign{},
		&ds.PancreatitisSignLike{},
	)
	if err != nil {
		log.Fatalf("не удалось выполнить миграцию: %v", err)
	}

	statements := []string{
		`ALTER TABLE pancreatitis_signs DROP CONSTRAINT IF EXISTS chk_pancreatitis_signs_status`,
		`ALTER TABLE pancreatitis_signs ADD CONSTRAINT chk_pancreatitis_signs_status CHECK (status IN ('черновик', 'опубликован', 'удален'))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_pancreatitis_signs_one_draft_per_user ON pancreatitis_signs (creator_id) WHERE status = 'черновик'`,
	}
	for _, statement := range statements {
		if err = db.Exec(statement).Error; err != nil {
			log.Fatalf("не удалось применить ограничение: %v\n%s", err, statement)
		}
	}

	log.Println("миграция выполнена успешно")
}
