package user

type User struct {
	ID    uint32
	Login string
}

func (u User) GetID() uint32 {
	return u.ID
}
