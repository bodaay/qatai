package db

type User struct {
	ID            string   `json:"id"`
	Email         string   `json:"email"`
	Name          string   `json:"name"`
	Salt          string   `json:"salt"`
	Password      string   `json:"password"` // This should be hashed & salted in a real-world scenario
	Disabled      bool     `json:"disabled"`
	LastLoginIP   string   `json:"last_login_ip"`
	LastLoginTime string   `json:"last_login_time"`
	ChatsIDs      []string `json:"chats"`
}

func AddUpdateUser(db *QataiDatabase, user *User, updateIfExists bool) error {

	return nil
}
