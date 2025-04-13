package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CustomerAddress struct {
    AddressType   string `bson:"addressType" json:"addressType"`
    IsDefault     bool   `bson:"isDefault" json:"isDefault"`
    AddressLine1  string `bson:"addressLine1" json:"addressLine1"`
    AddressLine2  string `bson:"addressLine2" json:"addressLine2"`
    City          string `bson:"city" json:"city"`
    State         string `bson:"state" json:"state"`
    PostalCode    string `bson:"postalCode" json:"postalCode"`
    Country       string `bson:"country" json:"country"`
}

type OrderHistoryEntry struct {
    OrderID     string    `bson:"orderId" json:"orderId"`
    OrderNumber string    `bson:"orderNumber" json:"orderNumber"`
    Date        time.Time `bson:"date" json:"date"`
    TotalAmount float64   `bson:"totalAmount" json:"totalAmount"` // Decimal128 → float64
    Status      string    `bson:"status" json:"status"`
}

type Customer struct {
    ID           primitive.ObjectID  `bson:"_id,omitempty" json:"_id"`
    CustomerID   string              `bson:"customerId" json:"customerId"`
    Email        string              `bson:"email" json:"email"`
    FirstName    string              `bson:"firstName" json:"firstName"`
    LastName     string              `bson:"lastName" json:"lastName"`
    Phone        string              `bson:"phone" json:"phone"`
    Addresses    []CustomerAddress   `bson:"addresses" json:"addresses"`
    OrderHistory []OrderHistoryEntry `bson:"orderHistory" json:"orderHistory"`
    Created      time.Time           `bson:"created" json:"created"`
    Updated      time.Time           `bson:"updated" json:"updated"`
}
