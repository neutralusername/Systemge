package SessionManager

type Identity struct {
	id       string
	sessions map[string]*Session
}

func (identity *Identity) GetId() string {
	return identity.id
}
