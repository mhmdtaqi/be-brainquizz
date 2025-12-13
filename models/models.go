package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	Name           string     `json:"name"`
	Email          string     `json:"email" gorm:"unique"`
	Password       []byte     `json:"-"`
	Role           string     `json:"role"`
	FailedAttempts int        `json:"failed_attempts" gorm:"default:0"`
	LockedUntil    *time.Time `json:"locked_until"`
}
type Kategori_Soal struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Tingkatan struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Kelas struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	JoinCode    string `json:"join_code" gorm:"unique"`
	CreatedBy   uint   `json:"created_by"`
	Creator     Users  `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
}
type Kuis struct {
	gorm.Model
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	IsPrivate     bool          `json:"is_private" gorm:"default:false"`
	Kategori_id   uint          `json:"kategori_id"`
	Kategori      Kategori_Soal `gorm:"foreignKey:Kategori_id;constraint:OnDelete:CASCADE;"`
	Tingkatan_id  uint          `json:"tingkatan_id"`
	Tingkatan     Tingkatan     `gorm:"foreignKey:Tingkatan_id;constraint:OnDelete:CASCADE;"`
	Kelas_id      uint          `json:"kelas_id"`
	Kelas         Kelas         `gorm:"foreignKey:Kelas_id;constraint:OnDelete:CASCADE;"`
	Pendidikan_id uint          `json:"pendidikan_id"`
	Pendidikan    Pendidikan    `gorm:"foreignKey:Pendidikan_id;constraint:OnDelete:CASCADE;"`
	CreatedBy     uint          `json:"created_by"`
	Creator       Users         `gorm:"foreignKey:CreatedBy;constraint:OnDelete:CASCADE;"`
}

type Soal struct {
	gorm.Model
	Question       string          `json:"question"`
	Options        json.RawMessage `json:"options_json"`
	Correct_answer string          `json:"correct_answer"`
	Kuis_id        uint            `json:"kuis_id"`
	Kuis           Kuis            `gorm:"foreignKey:Kuis_id;constraint:OnDelete:CASCADE;"`
}

type Pendidikan struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}
type Hasil_Kuis struct {
	gorm.Model
	Users_id       uint  `json:"users_id"`
	Users          Users `gorm:"foreignKey:Users_id;constraint:OnDelete:CASCADE;"`
	Kuis_id        uint  `json:"kuis_id"`
	Kuis           Kuis  `gorm:"foreignKey:Kuis_id;constraint:OnDelete:CASCADE;"`
	Score          uint  `json:"score"`
	Correct_Answer uint  `json:"correct_answer"`
}
type SoalAnswer struct {
	gorm.Model
	Soal_id uint   `json:"soal_id"`
	Soal    Soal   `gorm:"foreignKey:Soal_id;constraint:OnDelete:CASCADE;"`
	Answer  string `json:"answer"`
	User_id uint   `json:"user_id"`
	User    Users  `gorm:"foreignKey:User_id;constraint:OnDelete:CASCADE;"`
}
type Kelas_Pengguna struct {
	gorm.Model
	Users_id uint  `json:"users_id"`
	Users    Users `gorm:"foreignKey:Users_id;constraint:OnDelete:CASCADE;"`
	Kelas_id uint  `json:"kelas_id"`
	Kelas    Kelas `gorm:"foreignKey:Kelas_id;constraint:OnDelete:CASCADE;"`
}

type AuditLog struct {
	gorm.Model
	UserID    uint   `json:"user_id"`
	User      Users  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE;"`
	Action    string `json:"action"` // login, logout, failed_login
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}
