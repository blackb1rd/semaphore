package sql

import "github.com/semaphoreui/semaphore/db"

func (d *SqlDb) GetGlobalRole(roleID int) (db.Role, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where id=?", roleID)
	return role, err
}

func (d *SqlDb) GetProjectRoles(projectID int) ([]db.Role, error) {
	var roles []db.Role
	_, err := d.selectAll(&roles, "select * from `role` where project_id=? order by name", projectID)
	return roles, err
}

func (d *SqlDb) GetGlobalRoles() ([]db.Role, error) {
	var roles []db.Role
	_, err := d.selectAll(&roles, "select * from `role` where project_id is null order by name")
	return roles, err
}

func (d *SqlDb) UpdateRole(role db.Role) error {
	_, err := d.exec(
		"update `role` set slug=?, name=?, permissions=? where id=?",
		role.Slug,
		role.Name,
		role.Permissions,
		role.ID)
	return err
}

func (d *SqlDb) CreateRole(role db.Role) (db.Role, error) {
	insertID, err := d.insert(
		"id",
		"insert into `role` (slug, name, permissions, project_id) values (?, ?, ?, ?)",
		role.Slug,
		role.Name,
		role.Permissions,
		role.ProjectID)

	if err != nil {
		return role, err
	}

	role.ID = insertID
	return role, nil
}

func (d *SqlDb) DeleteRole(roleID int) error {
	res, err := d.exec("delete from `role` where id=?", roleID)
	return validateMutationResult(res, err)
}

func (d *SqlDb) GetProjectRole(projectID int, roleID int) (db.Role, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where id=? and project_id=?", roleID, projectID)
	return role, err
}

func (d *SqlDb) GetProjectOrGlobalRoleBySlug(projectID int, slug string) (db.Role, error) {
	var role db.Role
	err := d.selectOne(&role, "select * from `role` where slug=? and project_id=?", slug, projectID)
	return role, err
}
