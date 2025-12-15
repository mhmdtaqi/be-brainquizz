package database

import (
	"fmt"

	"github.com/Joko206/UAS_PWEB1/models"
)

// CreateKuis creates a new Kuis in the database
func CreateKuis(title string, description string, isPrivate bool, kategori uint, tingkatan uint, kelas uint, pendidikan uint, createdBy uint) (models.Kuis, error) {
	var newKuis = models.Kuis{
		Title:         title,
		Description:   description,
		IsPrivate:     isPrivate,
		Kategori_id:   kategori,
		Tingkatan_id:  tingkatan,
		Kelas_id:      kelas,
		Pendidikan_id: pendidikan,
		CreatedBy:     createdBy,
	}

	// Get DB connection
	db, err := GetDBConnection()
	if err != nil {
		return newKuis, err
	}

	// Validate Kategori, Tingkatan, and Kelas
	var kategoriObj models.Kategori_Soal
	if err := db.First(&kategoriObj, kategori).Error; err != nil {
		return newKuis, fmt.Errorf("invalid kategori id")
	}

	var tingkatanObj models.Tingkatan
	if err := db.First(&tingkatanObj, tingkatan).Error; err != nil {
		return newKuis, fmt.Errorf("invalid tingkatan id")
	}

	var kelasObj models.Kelas
	if err := db.First(&kelasObj, kelas).Error; err != nil {
		return newKuis, fmt.Errorf("invalid kelas id")
	}

	var PendidikanObj models.Pendidikan
	if err := db.First(&PendidikanObj, pendidikan).Error; err != nil {
		return newKuis, fmt.Errorf("invalid pendidikan id")
	}

	// Insert the new Kuis into the database
	if err := db.Create(&newKuis).Error; err != nil {
		return newKuis, fmt.Errorf("failed to insert data into kuis: %w", err)
	}

	return newKuis, nil
}

// GetKuis retrieves all Kuis from the database with related Kategori, Tingkatan, and Kelas
func GetKuis() ([]models.Kuis, error) {
	var kuisList []models.Kuis

	// Get DB connection
	db, err := GetDBConnection()
	if err != nil {
		return kuisList, err
	}

	// Preload related models (Kategori, Tingkatan, Kelas, Pendidikan)
	if err := db.Preload("Kategori").Preload("Tingkatan").Preload("Kelas").Preload("Pendidikan").Find(&kuisList).Error; err != nil {
		return kuisList, fmt.Errorf("failed to retrieve kuis: %w", err)
	}

	return kuisList, nil
}

// GetKuisForUser retrieves kuis that user can access (public + private from joined classes)
func GetKuisForUser(userID uint) ([]models.Kuis, error) {
	var kuisList []models.Kuis

	// Get DB connection
	db, err := GetDBConnection()
	if err != nil {
		return kuisList, err
	}

	// Get user's joined classes
	var userClasses []models.Kelas_Pengguna
	if err := db.Where("users_id = ?", userID).Find(&userClasses).Error; err != nil {
		return kuisList, fmt.Errorf("failed to get user classes: %w", err)
	}

	// Extract class IDs
	var classIDs []uint
	for _, uc := range userClasses {
		classIDs = append(classIDs, uc.Kelas_id)
	}

	// Query for accessible kuis:
	// 1. All public kuis (is_private = false)
	// 2. Private kuis from user's joined classes
	query := db.Preload("Kategori").Preload("Tingkatan").Preload("Kelas").Preload("Pendidikan")

	if len(classIDs) > 0 {
		// User has joined classes - can see public kuis + private kuis from joined classes
		query = query.Where("is_private = ? OR (is_private = ? AND kelas_id IN ?)", false, true, classIDs)
	} else {
		// User hasn't joined any class - can only see public kuis
		query = query.Where("is_private = ?", false)
	}

	if err := query.Find(&kuisList).Error; err != nil {
		return kuisList, fmt.Errorf("failed to retrieve accessible kuis: %w", err)
	}

	return kuisList, nil
}

// UpdateKuis updates an existing Kuis in the database
func UpdateKuis(title string, description string, isPrivate bool, kategori uint, tingkatan uint, kelas uint, pendidikan uint, id string) (models.Kuis, error) {
	var updatedKuis = models.Kuis{
		Title:         title,
		Description:   description,
		IsPrivate:     isPrivate,
		Kategori_id:   kategori,
		Tingkatan_id:  tingkatan,
		Kelas_id:      kelas,
		Pendidikan_id: pendidikan,
	}

	// Get DB connection
	db, err := GetDBConnection()
	if err != nil {
		return updatedKuis, err
	}

	// Update the kuis details
	if err := db.Where("ID = ?", id).Updates(&updatedKuis).Error; err != nil {
		return updatedKuis, fmt.Errorf("failed to update kuis: %w", err)
	}

	return updatedKuis, nil
}

// DeleteKuis deletes a Kuis by its ID
func DeleteKuis(id string) error {
	var kuis models.Kuis

	// Get DB connection
	db, err := GetDBConnection()
	if err != nil {
		return err
	}

	// Delete the Kuis by ID
	if err := db.Where("ID = ?", id).Delete(&kuis).Error; err != nil {
		return fmt.Errorf("failed to delete kuis: %w", err)
	}

	return nil
}
