package repository

import (
	"database/sql"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	err := r.db.QueryRow(
		"INSERT INTO users (name, is_active) VALUES ($1, $2) RETURNING id",
		user.Name, user.IsActive,
	).Scan(&user.ID)
	return err
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRow(
		"SELECT id, name, is_active FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Name, &user.IsActive)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (r *UserRepository) GetAll() ([]models.User, error) {
	rows, err := r.db.Query("SELECT id, name, is_active FROM users ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (r *UserRepository) Update(user *models.User) error {
	_, err := r.db.Exec(
		"UPDATE users SET name = $1, is_active = $2 WHERE id = $3",
		user.Name, user.IsActive, user.ID,
	)
	return err
}

func (r *UserRepository) GetActiveUsersByTeam(teamName string, excludeUserID int) ([]models.User, error) {
	query := `
		SELECT u.id, u.name, u.is_active 
		FROM users u
		INNER JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_name = $1 AND u.is_active = true AND u.id != $2
		ORDER BY RANDOM()
	`
	rows, err := r.db.Query(query, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// BulkDeactivateByTeam деактивирует всех пользователей команды
// Возвращает количество деактивированных пользователей
func (r *UserRepository) BulkDeactivateByTeam(teamName string) (int, error) {
	result, err := r.db.Exec(`
		UPDATE users 
		SET is_active = false 
		WHERE id IN (
			SELECT user_id FROM team_members WHERE team_name = $1
		) AND is_active = true
	`, teamName)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
