package enums

type Role int

const (
	Admin Role = iota + 1
	Teacher
	Student
)

func (r Role) ToString() string {
	switch r {
	case Admin:
		return "admin"
	case Teacher:
		return "teacher"
	case Student:
		return "student"
	default:
		return "Unknown"
	}
}
