package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ParentCategory struct {
    ID   string `bson:"id" json:"id"`
    Name string `bson:"name" json:"name"`
}

type Category struct {
    ID             string         `bson:"id" json:"id"`
    Name           string         `bson:"name" json:"name"`
    ParentCategory ParentCategory `bson:"parentCategory" json:"parentCategory"`
}

type Attribute struct {
    Name  string `bson:"name" json:"name"`
    Value string `bson:"value" json:"value"`
}

type Product struct {
    ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
    ProductID        string             `bson:"productId" json:"productId"`
    SKU              string             `bson:"sku" json:"sku"`
    Name             string             `bson:"name" json:"name"`
    Description      string             `bson:"description" json:"description"`
    Price            float64            `bson:"price" json:"price"`
    Category         Category           `bson:"category" json:"category"`
    CurrentInventory int                `bson:"currentInventory" json:"currentInventory"`
    Images           []string           `bson:"images" json:"images"`
    Attributes       []Attribute        `bson:"attributes" json:"attributes"`
    Created          time.Time          `bson:"created" json:"created"`
    Updated          time.Time          `bson:"updated" json:"updated"`
}
