package persistence

import (
	"context"
	"errors"

	"github.com/Timos-API/transformer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VueconPersistor struct {
	c *mongo.Collection
}

type Vuecon struct {
	VueconID     primitive.ObjectID `json:"-" bson:"_id" `
	Name         string             `json:"name" bson:"name" keep:"insert" validate:"required,gt=2"`
	LastModified int64              `json:"last_modified" bson:"last_modified" keep:"insert"`
	Size         int64              `json:"size" bson:"size" keep:"insert"`
	Src          string             `json:"src" bson:"src" keep:"insert"`
	Tags         []string           `json:"tags" bson:"tags" keep:"update,insert,omitempty"`
	Categories   []string           `json:"categories" bson:"categories" keep:"update,insert,omitempty"`
}

func NewVueconPersistor(c *mongo.Collection) *VueconPersistor {
	return &VueconPersistor{c}
}

func (p *VueconPersistor) Create(ctx context.Context, vuecon Vuecon) (*Vuecon, error) {

	p.Delete(ctx, vuecon.Name)

	cleaned := transformer.Clean(vuecon, "insert")
	res, err := p.c.InsertOne(ctx, cleaned)

	if err != nil {
		return nil, err
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return p.GetById(ctx, oid.Hex())
	}

	return nil, errors.New("something ubiquitous happened")
}

func (p *VueconPersistor) Update(ctx context.Context, name string, update interface{}) (*Vuecon, error) {

	res := p.c.FindOneAndUpdate(ctx, bson.M{"name": name}, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(options.After))

	if res.Err() != nil {
		return nil, res.Err()
	}

	var vuecon Vuecon
	err := res.Decode(&vuecon)

	if err != nil {
		return nil, err
	}

	return &vuecon, nil
}

func (p *VueconPersistor) Delete(ctx context.Context, name string) (bool, error) {

	res, err := p.c.DeleteOne(ctx, bson.M{"name": name})

	if err != nil {
		return false, err
	}

	if res.DeletedCount == 0 {
		return false, errors.New("nothing has been deleted")
	}

	return true, nil
}

func (p *VueconPersistor) GetById(ctx context.Context, name string) (*Vuecon, error) {

	res := p.c.FindOne(ctx, bson.M{"name": name})

	if res.Err() != nil {
		return nil, res.Err()
	}

	var vuecon Vuecon
	err := res.Decode(&vuecon)

	if err != nil {
		return nil, err
	}

	return &vuecon, nil
}

func (p *VueconPersistor) GetAll(ctx context.Context) (*[]Vuecon, error) {
	cursor, err := p.c.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"name": 1}))

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)
	vuecons := []Vuecon{}

	for cursor.Next(ctx) {
		var vuecon Vuecon
		err := cursor.Decode(&vuecon)

		if err != nil {
			return nil, err
		}

		vuecons = append(vuecons, vuecon)
	}

	if cursor.Err() != nil {
		return nil, cursor.Err()
	}

	return &vuecons, nil
}
