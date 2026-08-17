package user

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
)

// mongoUser คือโครงสร้างข้อมูลที่ใช้สำหรับเก็บ user ลงใน mongoDB
// แยกออกจาก domainuser.User เผื่ออยาคตอยากย้ายไปใช้ database ตัวอื่น
// โดยที่ domain layer ไม่ต้องรู้จักกับ mongoDB
type mongoUser struct {
	ID        string    `bson:"_id,omitempty"`
	Name      string    `bson:"name"`
	Email     string    `bson:"email"`
	Password  string    `bson:"password"`
	CreatedAt time.Time `bson:"created_at"`
}

// repository เป็น mongodb Adapter สำหรับจัดการข้อมูล User
// โดย implement domainuser.Repository interface
type Repository struct {
	collection *mongo.Collection
}

func New(collection *mongo.Collection) *Repository {
	return &Repository{collection: collection}
}

// toDomain แปลงข้อมูลจาก mongoDB model
// กลับไปเป็น domain model ที่ application/domain ใช้งาน
func toDomain(m mongoUser) *domainuser.User {
	return &domainuser.User{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Password:  m.Password,
		CreatedAt: m.CreatedAt,
	}
}

func (r *Repository) Create(ctx context.Context, u *domainuser.User) error {
	doc := mongoUser{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
	}
	_, err := r.collection.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return domainuser.ErrUsernameAlreadyExists
	}
	return err
}

func (r *Repository) SearchByID(ctx context.Context, id string) (*domainuser.User, error) {
	var m mongoUser
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domainuser.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(m), nil
}

func (r *Repository) SearchByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var m mongoUser
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&m)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domainuser.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(m), nil
}

func (r *Repository) List(ctx context.Context) ([]*domainuser.User, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domainuser.User
	for cursor.Next(ctx) {
		var m mongoUser
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		users = append(users, toDomain(m))
	}
	return users, cursor.Err()
}

func (r *Repository) Update(ctx context.Context, id string, req domainuser.UpdateRequest) (*domainuser.User, error) {
	update := bson.M{}
	if req.Name != nil {
		update["name"] = *req.Name
	}
	if req.Email != nil {
		update["email"] = *req.Email
	}
	if len(update) == 0 {
		return r.SearchByID(ctx, id)
	}

	// FindOneAndUpdate ใช้ค้นหา user จาก ID
	// แล้ว update เฉพาะ field ที่อยู่ใน $set
	result := r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	)

	var m mongoUser
	if err := result.Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domainuser.ErrUserNotFound
		}
		return nil, err
	}
	return r.SearchByID(ctx, id) // return the post-update document
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return domainuser.ErrUserNotFound
	}
	return nil
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
