package Tools

import "sync"

type AccessControlList_ struct {
	list  map[string]bool
	mutex sync.Mutex
}

func NewAccessControlList() *AccessControlList_ {
	return &AccessControlList_{
		list: make(map[string]bool),
	}
}

func (acl *AccessControlList_) Add(item string) {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	acl.list[item] = true
}

func (acl *AccessControlList_) Remove(item string) {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	delete(acl.list, item)
}

func (acl *AccessControlList_) Contains(item string) bool {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	_, ok := acl.list[item]
	return ok
}

func (acl *AccessControlList_) ElementCount() int {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	return len(acl.list)
}

func (acl *AccessControlList_) GetElements() []string {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	items := make([]string, 0, len(acl.list))
	for item := range acl.list {
		items = append(items, item)
	}
	return items
}
