package models

import "time"

type Role struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100;not null;unique"`
}

func (Role) TableName() string { return "roles" }

type User struct {
	ID             uint   `gorm:"primaryKey"`
	Login          string `gorm:"size:50;not null;unique"`
	PasswordHash   string `gorm:"size:255;not null"`
	LastName       string `gorm:"size:100;not null"`
	FirstName      string `gorm:"size:100;not null"`
	MiddleName     string `gorm:"size:100"`
	RoleID         uint
	Role           Role   `gorm:"foreignKey:RoleID"`
	ContactNumber  string `gorm:"size:20"`
	ContractNumber string `gorm:"size:50;not null"`
}

func (User) TableName() string { return "users" }

type WindTunnel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:50;not null;unique"`
}

func (WindTunnel) TableName() string { return "wind_tunnels" }

type ModelLA struct {
	ID          uint   `gorm:"primaryKey"`
	ModelNumber string `gorm:"size:50;not null;unique"`
	CodeName    string `gorm:"size:100;not null"`
}

// Силовое исправление для GORM, чтобы он видел твою таблицу model_la
func (ModelLA) TableName() string { return "models" }

type Equipment struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:150;not null"`
	CodeName string `gorm:"size:50;not null;unique"`
}

func (Equipment) TableName() string { return "equipment" }

type ExperimentType struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:100;not null;unique"`
}

func (ExperimentType) TableName() string { return "experiment_types" }

type Experiment struct {
	ID               uint   `gorm:"primaryKey"`
	ExperimentNumber string `gorm:"size:50;not null;unique"`
	ExperimentName   string `gorm:"size:250;not null"`
	ContractNumber   string `gorm:"size:50;not null"`
	WindTunnelID     uint
	WindTunnel       WindTunnel `gorm:"foreignKey:WindTunnelID"`
	ModelID          uint
	ModelLA          ModelLA `gorm:"foreignKey:ModelID"`
	StartDate        time.Time
	EndDate          *time.Time
	TunnelChiefID    uint
	TunnelChief      User `gorm:"foreignKey:TunnelChiefID"`
	LeadEngineerID   uint
	LeadEngineer     User `gorm:"foreignKey:LeadEngineerID"`
}

func (Experiment) TableName() string { return "experiments" }

// Та самая 12-я промежуточная таблица для оборудования в экспериментах
type ExperimentEquipment struct {
	ExperimentID uint `gorm:"primaryKey"`
	EquipmentID  uint `gorm:"primaryKey"`
}

func (ExperimentEquipment) TableName() string { return "experiment_equipment" }

type Shift struct {
	ID              uint `gorm:"primaryKey"`
	ExperimentID    uint
	Experiment      Experiment `gorm:"foreignKey:ExperimentID"`
	ShiftNumber     int        `gorm:"not null"`
	BrigadierID     uint
	Brigadier       User      `gorm:"foreignKey:BrigadierID"`
	WorkDescription string    `gorm:"type:text;not null"`
	ShiftDate       time.Time `gorm:"default:CURRENT_DATE"`
}

func (Shift) TableName() string { return "shifts" }

type Configuration struct {
	ID          uint `gorm:"primaryKey"`
	ShiftID     uint
	Shift       Shift `gorm:"foreignKey:ShiftID"`
	Description string
	WindSpeed   float64
	RollAngle   float64
	YawAngle    float64
}

func (Configuration) TableName() string { return "configurations" }

type Protocol struct {
	ID                uint `gorm:"primaryKey"`
	ConfigurationID   uint
	Configuration     Configuration `gorm:"foreignKey:ConfigurationID"`
	ProtocolNumber    string        `gorm:"size:50;not null"`
	VariableParameter string        `gorm:"size:50"`
}

func (Protocol) TableName() string { return "protocols" }

type ProtocolData struct {
	ID         uint64 `gorm:"primaryKey"`
	ProtocolID uint
	Protocol   Protocol `gorm:"foreignKey:ProtocolID"`
	PointN     string   `gorm:"size:10"`
	Al         float64
	Alpha      float64
	Beta       float64
	Q          float64
	V          float64
	Pf         float64
	Pa         float64
	Tf         float64
	X          float64
	Y          float64
	Z          float64
	Mx         float64
	My         float64
	Mz         float64
}

func (ProtocolData) TableName() string { return "protocol_data" }
