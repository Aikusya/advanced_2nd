package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Person struct {
	ID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	FullName   string             `json:"fullname"`
	BirthDate  string             `json:"birthDate"`
	Address    Address            `json:"address"`
	Contacts   []string           `json:"contacts"`
	IsStudent  bool               `json:"isStudent"`
	IsEmployed bool               `json:"isEmployed"`
	IsChecked  bool               `json: isChecked`
}

type Address struct {
	City    string `json:"city"`
	Country string `json:"country"`
}

type JsonResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
