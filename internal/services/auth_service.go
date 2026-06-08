package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sigap2/sigap2/internal/config"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(name, email, password, role, phone string) error {
	// check if email exists
	var existing models.User
	database.DB.Where("email = ?", email).First(&existing)
	if existing.ID != 0 {
		return errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := models.User{
		Name:          name,
		Email:         email,
		PasswordHash:  string(hash),
		PlainPassword: password,
		Role:          role,
		Phone:         phone,
	}

	return database.DB.Create(&user).Error
}

func LoginUser(email, password string) (string, error) {
	var user models.User
	database.DB.Where("email = ?", email).First(&user)

	if user.ID == 0 {
		return "", errors.New("invalid credentials")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func UpdateUserPassword(userID uint, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return database.DB.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"password_hash":  string(hash),
		"plain_password": newPassword,
	}).Error
}

func ChangeOwnPassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("password lama tidak sesuai")
	}

	return UpdateUserPassword(userID, newPassword)
}
