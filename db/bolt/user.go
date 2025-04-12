package bolt

import (
	"fmt"
	"github.com/semaphoreui/semaphore/db"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func (d *BoltDb) CreateUserWithoutPassword(user db.User) (newUser db.User, err error) {

	err = db.ValidateUser(user)
	if err != nil {
		return
	}

	_, err = d.GetUserByLoginOrEmail(user.Username, user.Email)

	if err == nil {
		err = fmt.Errorf("user already exists")
		return
	}

	if err != db.ErrNotFound {
		return
	}

	user.Password = ""
	user.Created = db.GetParsedTime(time.Now().UTC())

	usr, err := d.createObject(0, db.UserProps, user)

	if err != nil {
		return
	}

	newUser = usr.(db.User)
	return
}

func (d *BoltDb) CreateUser(user db.UserWithPwd) (newUser db.User, err error) {

	err = db.ValidateUser(user.User)
	if err != nil {
		return
	}

	_, err = d.GetUserByLoginOrEmail(user.Username, user.Email)

	if err == nil {
		err = fmt.Errorf("user already exists")
		return
	}

	if err != db.ErrNotFound {
		return
	}

	pwdHash, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), 11)

	if err != nil {
		return
	}

	user.Password = string(pwdHash)
	user.Created = db.GetParsedTime(time.Now().UTC())

	usr, err := d.createObject(0, db.UserProps, user)

	if err != nil {
		return
	}

	newUser = usr.(db.UserWithPwd).User
	return
}

func (d *BoltDb) DeleteUser(userID int) error {
	projects, err := d.GetProjects(userID)
	if err != nil {
		return err
	}

	// TODO: add transaction

	for _, p := range projects {
		_ = d.DeleteProjectUser(p.ID, userID)
	}

	return d.deleteObject(0, db.UserProps, intObjectID(userID), nil)
}

func (d *BoltDb) UpdateUser(user db.UserWithPwd) error {
	var password string

	if user.Pwd != "" {
		var pwdHash []byte
		pwdHash, err := bcrypt.GenerateFromPassword([]byte(user.Pwd), 11)
		if err != nil {
			return err
		}
		password = string(pwdHash)
	} else {
		oldUser, err := d.GetUser(user.ID)
		if err != nil {
			return err
		}
		password = oldUser.Password
	}

	user.Password = password

	return d.updateObject(0, db.UserProps, user)
}

func (d *BoltDb) SetUserPassword(userID int, password string) error {
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	if err != nil {
		return err
	}
	user, err := d.GetUser(userID)
	if err != nil {
		return err
	}
	user.Password = string(pwdHash)
	return d.updateObject(0, db.UserProps, user)
}

func (d *BoltDb) CreateProjectUser(projectUser db.ProjectUser) (db.ProjectUser, error) {
	newProjectUser, err := d.createObject(projectUser.ProjectID, db.ProjectUserProps, projectUser)

	if err != nil {
		return db.ProjectUser{}, err
	}

	return newProjectUser.(db.ProjectUser), nil
}

func (d *BoltDb) GetProjectUser(projectID, userID int) (user db.ProjectUser, err error) {
	err = d.getObject(projectID, db.ProjectUserProps, intObjectID(userID), &user)
	return
}

func (d *BoltDb) GetProjectUsers(projectID int, params db.RetrieveQueryParams) (users []db.UserWithProjectRole, err error) {
	var projectUsers []db.ProjectUser
	err = d.getObjects(projectID, db.ProjectUserProps, params, nil, &projectUsers)
	if err != nil {
		return
	}
	for _, projUser := range projectUsers {
		var usr db.User
		usr, err = d.GetUser(projUser.UserID)
		if err != nil {
			return
		}
		var usrWithRole = db.UserWithProjectRole{
			User: usr,
			Role: projUser.Role,
		}
		users = append(users, usrWithRole)
	}
	return
}

func (d *BoltDb) UpdateProjectUser(projectUser db.ProjectUser) error {
	return d.updateObject(projectUser.ProjectID, db.ProjectUserProps, projectUser)
}

func (d *BoltDb) DeleteProjectUser(projectID, userID int) error {
	return d.deleteObject(projectID, db.ProjectUserProps, intObjectID(userID), nil)
}

func (d *BoltDb) getTotp(userID int) (res *db.UserTotp, err error) {

	current := make([]db.UserTotp, 0)
	err = d.getObjects(userID, db.UserTotpProps, db.RetrieveQueryParams{}, nil, &current)

	if err != nil {
		return
	}

	if len(current) > 0 {
		res = &current[0]
	}

	return
}

// GetUser retrieves a user from the database by ID
func (d *BoltDb) GetUser(userID int) (user db.User, err error) {
	err = d.getObject(0, db.UserProps, intObjectID(userID), &user)
	if err != nil {
		return
	}

	user.Totp, err = d.getTotp(userID)

	return
}

func (d *BoltDb) GetProUserCount() (count int, err error) {
	var users []db.User
	err = d.getObjects(0, db.UserProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		user := i.(db.User)
		return user.Pro
	}, &users)
	if err != nil {
		return
	}
	count = len(users)
	return
}

func (d *BoltDb) GetUserCount() (count int, err error) {
	var users []db.User
	err = d.getObjects(0, db.UserProps, db.RetrieveQueryParams{}, nil, &users)
	if err != nil {
		return
	}
	count = len(users)
	return
}

func (d *BoltDb) GetUsers(params db.RetrieveQueryParams) (users []db.User, err error) {
	err = d.getObjects(0, db.UserProps, params, nil, &users)
	return
}

func (d *BoltDb) GetUserByLoginOrEmail(login string, email string) (existingUser db.User, err error) {
	var users []db.User
	err = d.getObjects(0, db.UserProps, db.RetrieveQueryParams{}, nil, &users)
	if err != nil {
		return
	}

	found := false

	for _, user := range users {
		if user.Username == login || user.Email == email {
			existingUser = user
			found = true
			break
		}
	}

	if !found {
		err = db.ErrNotFound
		return
	}

	existingUser.Totp, err = d.getTotp(existingUser.ID)

	return
}

func (d *BoltDb) GetAllAdmins() (users []db.User, err error) {
	err = d.getObjects(0, db.UserProps, db.RetrieveQueryParams{}, func(i interface{}) bool {
		user := i.(db.User)
		return user.Admin
	}, &users)
	return
}

func (d *BoltDb) AddTotpVerification(userID int, url string, recoveryHash string) (totp db.UserTotp, err error) {

	current := make([]db.UserTotp, 0)
	err = d.getObjects(userID, db.UserTotpProps, db.RetrieveQueryParams{}, nil, &current)

	if len(current) > 0 {
		err = fmt.Errorf("already exists")
		return
	}

	totp.UserID = userID
	totp.URL = url
	totp.RecoveryHash = recoveryHash
	totp.Created = db.GetParsedTime(time.Now().UTC())

	newTotp, err := d.createObject(userID, db.UserTotpProps, totp)

	if err != nil {
		return
	}

	totp = newTotp.(db.UserTotp)
	return
}

func (d *BoltDb) DeleteTotpVerification(userID int, totpID int) error {
	return d.deleteObject(userID, db.UserTotpProps, intObjectID(totpID), nil)
}
