package Tools

import "sync"

type AccessControlList struct {
	list  map[string]bool
	mutex sync.Mutex
}

func NewAccessControlList(entries []string) *AccessControlList {
	list := map[string]bool{}
	for _, item := range entries {
		list[item] = true
	}
	return &AccessControlList{
		list: list,
	}
}

func (acl *AccessControlList) Add(item string) {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	acl.list[item] = true
}

func (acl *AccessControlList) Remove(item string) {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	delete(acl.list, item)
}

func (acl *AccessControlList) Contains(item string) bool {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	_, ok := acl.list[item]
	return ok
}

func (acl *AccessControlList) ElementCount() int {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	return len(acl.list)
}

func (acl *AccessControlList) GetElements() []string {
	acl.mutex.Lock()
	defer acl.mutex.Unlock()
	items := make([]string, 0, len(acl.list))
	for item := range acl.list {
		items = append(items, item)
	}
	return items
}
