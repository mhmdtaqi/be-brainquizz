package controllers

import (
	"strconv"
	"time"

	"github.com/Joko206/UAS_PWEB1/database"
	"github.com/Joko206/UAS_PWEB1/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// Secret key for JWT
const SecretKey = "secret"

// Helper function to authenticate using JWT
func Authenticate(c *fiber.Ctx) (*models.Users, error) {
	var tokenString string

	// First try to get token from cookie
	cookie := c.Cookies("jwt")
	if cookie != "" {
		tokenString = cookie
	} else {
		// If no cookie, try Authorization header
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}
	}

	if tokenString == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "No JWT token found")
	}

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired token")
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["iss"].(string)
	var user models.Users
	database.DB.Where("id = ?", userID).First(&user)
	if user.ID == 0 {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "User not found")
	}

	return &user, nil
}

func RoleMiddleware(allowedRoles []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, err := Authenticate(c)
		if err != nil {
			return err
		}

		// Check if the user role is allowed
		roleAllowed := false
		for _, role := range allowedRoles {
			if user.Role == role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"data":    nil,
				"success": false,
				"message": "You don't have permission to access this resource",
			})
		}

		return c.Next()
	}
}

// LogAudit logs user activities for auditing
func LogAudit(userID uint, action string, c *fiber.Ctx) {
	ip := c.IP()
	userAgent := c.Get("User-Agent")

	auditLog := models.AuditLog{
		UserID:    userID,
		Action:    action,
		IPAddress: ip,
		UserAgent: userAgent,
	}

	database.DB.Create(&auditLog)
}

func Register(c *fiber.Ctx) error {
	var data map[string]string

	// Parse request body
	if err := c.BodyParser(&data); err != nil {
		return sendResponse(c, fiber.StatusBadRequest, false, "Invalid request body", nil)
	}

	// Default role is "student" if not provided
	role := data["role"]
	if role == "" {
		role = "student" // Default role
	}

	// Validate that the role is one of the allowed values
	if role != "admin" && role != "teacher" && role != "student" {
		return sendResponse(c, fiber.StatusBadRequest, false, "Invalid role. Allowed roles: admin, teacher, student", nil)
	}

	// Hash password before saving
	password, err := bcrypt.GenerateFromPassword([]byte(data["password"]), 14)
	if err != nil {
		return sendResponse(c, fiber.StatusInternalServerError, false, "Error hashing password", nil)
	}

	// Create user with the role
	user := models.Users{
		Name:     data["name"],
		Email:    data["email"],
		Password: password,
		Role:     role, // Set the role here
	}

	// Save user to the database
	if err := database.DB.Create(&user).Error; err != nil {
		return handleError(c, err, "Failed to register user")
	}

	// Return success response
	return sendResponse(c, fiber.StatusOK, true, "User registered successfully", user)
}

func Login(c *fiber.Ctx) error {
	var data map[string]string

	// Parse request body
	if err := c.BodyParser(&data); err != nil {
		return sendResponse(c, fiber.StatusBadRequest, false, "Invalid request body", nil)
	}

	var user models.Users
	// Find user by email
	if err := database.DB.Where("email = ?", data["email"]).First(&user).Error; err != nil {
		// Log failed login attempt for unknown user
		LogAudit(0, "failed_login_unknown_user", c)
		return sendResponse(c, fiber.StatusNotFound, false, "User not found", nil)
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		LogAudit(user.ID, "failed_login_locked", c)
		return sendResponse(c, fiber.StatusTooManyRequests, false, "Account is temporarily locked due to too many failed attempts", nil)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword(user.Password, []byte(data["password"])); err != nil {
		// Increment failed attempts
		user.FailedAttempts++
		if user.FailedAttempts >= 3 {
			// Lock account for 15 minutes
			lockTime := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &lockTime
			user.FailedAttempts = 0 // Reset after lock
		}
		database.DB.Save(&user)
		LogAudit(user.ID, "failed_login", c)
		return sendResponse(c, fiber.StatusUnauthorized, false, "Invalid password", nil)
	}

	// Successful login - reset failed attempts and unlock
	user.FailedAttempts = 0
	user.LockedUntil = nil
	database.DB.Save(&user)

	// Generate JWT token with custom claims including role
	claims := jwt.MapClaims{
		"iss":  strconv.Itoa(int(user.ID)),
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
		"role": user.Role,
		"name": user.Name,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return handleError(c, err, "Failed to generate token")
	}

	// Set JWT cookie
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	// Log successful login
	LogAudit(user.ID, "login", c)

	return sendResponse(c, fiber.StatusOK, true, "Login successful", fiber.Map{
		"token":   tokenString,
		"role":    user.Role,
		"user_id": user.ID,
		"name":    user.Name,
	})
}

func User(c *fiber.Ctx) error {
	// Authenticate the user using the JWT token
	user, err := Authenticate(c)
	if err != nil {
		return err
	}

	// Return user details
	return sendResponse(c, fiber.StatusOK, true, "User retrieved successfully", user)
}

func Logout(c *fiber.Ctx) error {
	// Try to get user for logging (don't fail if token is invalid)
	if user, err := Authenticate(c); err == nil {
		LogAudit(user.ID, "logout", c)
	}

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour * 24),
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	// Return success message
	return sendResponse(c, fiber.StatusOK, true, "Logout successful", nil)
}
