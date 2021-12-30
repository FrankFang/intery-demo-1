package model

type Deployment struct {
	BaseModel
	ContainerId string `json:"container_id" gorm:"not null"`
	ProjectId   uint `json:"project_id" gorm:"not null"`
}
