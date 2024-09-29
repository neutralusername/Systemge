package Tools

type SessionManager struct {
	sessions   map[string]*Session
	identities map[string]*Identity

	onExpire func(*Session)
	onCreate func(*Session)
}

type Session struct {
	id       string
	identity *Identity
}

type Identity struct {
	id       string
	sessions map[string]*Session
}
