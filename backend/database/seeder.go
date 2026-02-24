package database

import (
	"log"

	"weekly-dashboard/models"
)

// Seed populates the database with initial data
func Seed() error {
	log.Println("Seeding database...")

	// Seed indicators
	if err := seedIndicators(); err != nil {
		return err
	}

	log.Println("Database seeding completed successfully")
	return nil
}

func seedIndicators() error {
	indicators := models.GetDefaultIndicators()

	for _, indicator := range indicators {
		// Check if indicator already exists
		var existing models.Indicator
		result := DB.Where("code = ?", indicator.Code).First(&existing)

		if result.RowsAffected == 0 {
			// Create new indicator
			if err := DB.Create(&indicator).Error; err != nil {
				log.Printf("Failed to seed indicator %s: %v", indicator.Code, err)
				return err
			}
			log.Printf("Seeded indicator: %s - %s", indicator.Code, indicator.Name)
		} else {
			// Update existing indicator
			existing.Department = indicator.Department
			existing.Name = indicator.Name
			existing.UnitOfMeasure = indicator.UnitOfMeasure
			existing.SpreadsheetName = indicator.SpreadsheetName
			existing.SpreadsheetRow = indicator.SpreadsheetRow
			existing.IsInverse = indicator.IsInverse
			existing.DisplayOrder = indicator.DisplayOrder
			existing.IsActive = indicator.IsActive

			if err := DB.Save(&existing).Error; err != nil {
				log.Printf("Failed to update indicator %s: %v", indicator.Code, err)
				return err
			}
			log.Printf("Updated indicator: %s - %s (SpreadsheetName='%s')", indicator.Code, indicator.Name, existing.SpreadsheetName)
		}
	}

	return nil
}
