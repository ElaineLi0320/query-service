package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItem struct {
    ProductID   string  `bson:"productId" json:"productId"`
    ProductName string  `bson:"productName" json:"productName"`
    SKU         string  `bson:"sku" json:"sku"`
    Quantity    int     `bson:"quantity" json:"quantity"`
    UnitPrice   float64 `bson:"unitPrice" json:"unitPrice"`
    TotalPrice  float64 `bson:"totalPrice" json:"totalPrice"`
}

type ShippingAddress struct {
    AddressLine1 string `bson:"addressLine1" json:"addressLine1"`
    AddressLine2 string `bson:"addressLine2" json:"addressLine2"`
    City         string `bson:"city" json:"city"`
    State        string `bson:"state" json:"state"`
    PostalCode   string `bson:"postalCode" json:"postalCode"`
    Country      string `bson:"country" json:"country"`
}

type Order struct {
    ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
    OrderID        string             `bson:"orderId" json:"orderId"`
    OrderNumber    string             `bson:"orderNumber" json:"orderNumber"`
    CustomerID     string             `bson:"customerId" json:"customerId"`
    CustomerEmail  string             `bson:"customerEmail" json:"customerEmail"`
    CustomerName   string             `bson:"customerName" json:"customerName"`
    Status         string             `bson:"status" json:"status"`
    TotalAmount    float64            `bson:"totalAmount" json:"totalAmount"`
    Items          []OrderItem        `bson:"items" json:"items"`
    ShippingAddress ShippingAddress   `bson:"shippingAddress" json:"shippingAddress"`
    Created        time.Time          `bson:"created" json:"created"`
    Updated        time.Time          `bson:"updated" json:"updated"`
}
