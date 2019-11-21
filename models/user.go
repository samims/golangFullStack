package models

import (
	"errors"
	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
	"html"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        uint32    `gorm:"primary_key; auto_increment" json:"id"`
	Nickname  string    `gorm:"size:255; not null; unique" json:"nickname"`
	Email     string    `gorm:"size:100; not null; unique" json:"email"`
	Password  string    `gorm:"size:100 not null;"json:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"json:"created_at"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"json:"updated_at"`
}

// Hash hashes the password
func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// VerifyPassword verifies password with hashed password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (u *User) BeforeSave() error {
	hashedPassword, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) Prepare() {
	u.ID = 0
	u.Nickname = html.EscapeString(strings.TrimSpace(u.Nickname))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
}

func (u *User) Validate(action string) error {
	switch strings.ToLower(action) {
	// code repetition
	case "update":
		if u.Nickname == "" {
			return errors.New("required Nickname")
		}
		if u.Password == "" {
			return errors.New("required Password")
		}
		if u.Email == "" {
			return errors.New("required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("invalid Email")
		}

		return nil
	case "login":
		if u.Password == "" {
			return errors.New("required Password")
		}
		if u.Email == "" {
			return errors.New("required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("invalid Email")
		}
		return nil

	default:
		if u.Nickname == "" {
			return errors.New("required Nickname")
		}
		if u.Password == "" {
			return errors.New("required Password")
		}
		if u.Email == "" {
			return errors.New("required Email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("invalid Email")
		}
		return nil
	}
}

func (u *User) SaveUser(db *gorm.DB) (*User, error) {
	var err error
	err = db.Debug().Create(&u).Error
	if err != nil {
		return &User{}, err
	}

	return u, nil
}

func (u *User) FindAllUsers(db *gorm.DB) (*[]User, error) {
	var err error
	var users []User
	err = db.Debug().Model(&User{}).Limit(100).Find(&u).Error
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindUserById(db *gorm.DB, uid uint32) (*User, error) {
	var err error
	err = db.Debug().Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("user not found")
	}
	return u, err
}

func (u *User) UpdateAUser(db *gorm.DB, uid uint32) (*User, error) {
	// To hash passwor
	err := u.BeforeSave()
	if err != nil {
		log.Fatal(err)
	}
	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":   u.Password,
			"nickname":   u.Nickname,
			"email":      u.Email,
			"updated_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &User{}, db.Error
	}

	// To show updated User
	err = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) DeleteAUser(db *gorm.DB, uid uint32) (int64, error) {
	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).Delete(&User{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil

}

type Post struct {
	ID        uint64    `gorm:"primary_key;auto_increment"json:"id"'`
	Title     string    `gorm:"size:255; not null; unique"json:"title"`
	Content   string    `gorm:"size:255;not null;unique"json:"content"`
	Author    User      `json:"author"`
	AuthorID  uint32    `gorm:"not null"json:"author_id"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"json:"created_at"`
	UpdatedAT time.Time `gorm:"default:CURRENT_TIMESTAMP"json:"updated_at"`
}

func (p *Post) Prepare() {
	p.ID = 0
	p.Title = html.EscapeString(strings.TrimSpace(p.Title))
	p.Content = html.EscapeString(strings.TrimSpace(p.Content))
	p.Author = User{}
	p.CreatedAt = time.Now()
	p.UpdatedAT = time.Now()

}

func (p *Post) Validate() error {
	if p.Title == "" {
		return errors.New("required Title")
	}
	if p.Content == "" {
		return errors.New("required Content")
	}
	if p.AuthorID < 1 {
		return errors.New("required author")
	}
	return nil

}

func (p *Post) SavePost(db *gorm.DB) (*Post, error) {
	var err error
	err = db.Debug().Model(&Post{}).Create(&p).Error
	if err != nil {
		return &Post{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Post{}, err
		}
	}
	return p, nil
}

func FindAllPosts(db *gorm.DB) (*[]Post, error) {
	var err error
	var posts []Post
	err = db.Debug().Model(&Post{}).Limit(100).Find(&posts).Error
	if err != nil {
		return &[]Post{}, err
	}
	if len(posts) > 0 {
		for i, _ := range posts {
			err := db.Debug().Model(&User{}).Where("id = ?", posts[i].AuthorID).Take(&posts[i].Author).Error
			if err != nil {
				return &[]Post, err
			}
		}
	}
	return &posts, nil

}

func (p *Post) FindPostByID(db *gorm.DB, pid int64) (*Post, error) {
	var err error
	err = db.Debug().Model(&Post{}).Where("id = ?", pid).Take(*p.Author).Error
	if err != nil {
		if err != nil {
			return &Post{}, err
		}
	}

	return p, nil
}

func (p *Post) UpdatePost(db *gorm.DB) (*Post, error) {
	var err error
	err = db.Debug().Model(&Post{}).Where("id = ?", p.ID).Updates(
		Post{
			Title:     p.Title,
			Content:   p.Content,
			UpdatedAT: time.Now(),
		}).Error
	if err != nil {
		return &Post{}, err
	}
	if p.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", p.AuthorID).Take(&p.Author).Error
		if err != nil {
			return &Post{}, err
		}
	}
	return p, nil
}

func (p *Post)  DeleteAPost(db *gorm.DB, pid uint64, uid uint32) (int64, error){
	db = db.Model(&Post{}).Where("id = ? and author_id = ?", pid, uid).Take(&Post{}).Delete(&Post{})
	if db.Error != nil{
		if gorm.IsRecordNotFoundError(db.Error){
			return 0, errors.New("post not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
