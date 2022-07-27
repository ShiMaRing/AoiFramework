package session

import "testing"

type User struct {
	Name string `aoiorm:"PRIMARY KEY"`
	Age  int
}

var (
	user1 = &User{"Tom", 18}
	user2 = &User{"Sam", 25}
	user3 = &User{"Jack", 25}
)

func TestRecordInit(t *testing.T) {
	t.Helper()
	s := NewSession().Model(&User{})
	err1 := s.DropTable()
	err2 := s.CreateTable()
	_, err3 := s.Insert(user1, user2)
	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("failed init test records")
	}
}

func TestSession_Insert(t *testing.T) {
	s := NewSession().Model(&User{})
	affected, err := s.Insert(user3)
	if err != nil || affected != 1 {
		t.Fatal("failed to create record")
	}
}

func TestSession_Find(t *testing.T) {
	s := NewSession().Model(&User{})
	var users []User
	if err := s.Find(&users); err != nil || len(users) != 3 {
		t.Log(err)
		t.Log(len(users))
		t.Fatal("failed to query all")

	}
}
