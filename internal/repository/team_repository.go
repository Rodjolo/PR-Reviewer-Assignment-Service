package repository

import (
	"database/sql"
	"pr-reviewer-service/internal/models"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(team *models.Team) error {
	_, err := r.db.Exec("INSERT INTO teams (name) VALUES ($1)", team.Name)
	return err
}

func (r *TeamRepository) GetByName(name string) (*models.Team, error) {
	team := &models.Team{Name: name}
	
	rows, err := r.db.Query(`
		SELECT u.id, u.name, u.is_active 
		FROM users u
		INNER JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_name = $1
		ORDER BY u.id
	`, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.IsActive); err != nil {
			return nil, err
		}
		members = append(members, user)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	// Проверяем существование команды
	var exists bool
	err = r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", name).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	
	team.Members = members
	return team, nil
}

func (r *TeamRepository) GetAll() ([]models.Team, error) {
	// Используем JOIN для загрузки всех данных за один запрос
	rows, err := r.db.Query(`
		SELECT t.name, u.id, u.name, u.is_active
		FROM teams t
		LEFT JOIN team_members tm ON t.name = tm.team_name
		LEFT JOIN users u ON tm.user_id = u.id
		ORDER BY t.name, u.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teamsMap := make(map[string]*models.Team)
	for rows.Next() {
		var teamName string
		var userID sql.NullInt64
		var userName sql.NullString
		var isActive sql.NullBool

		if err := rows.Scan(&teamName, &userID, &userName, &isActive); err != nil {
			return nil, err
		}

		// Создаем команду, если её еще нет
		if _, exists := teamsMap[teamName]; !exists {
			teamsMap[teamName] = &models.Team{
				Name:    teamName,
				Members: []models.User{},
			}
		}

		// Добавляем участника, если он есть
		if userID.Valid {
			teamsMap[teamName].Members = append(teamsMap[teamName].Members, models.User{
				ID:       int(userID.Int64),
				Name:     userName.String,
				IsActive: isActive.Bool,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Преобразуем map в slice
	teams := make([]models.Team, 0, len(teamsMap))
	for _, team := range teamsMap {
		teams = append(teams, *team)
	}

	return teams, nil
}

func (r *TeamRepository) getTeamMembers(teamName string) ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.name, u.is_active 
		FROM users u
		INNER JOIN team_members tm ON u.id = tm.user_id
		WHERE tm.team_name = $1
		ORDER BY u.id
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.IsActive); err != nil {
			return nil, err
		}
		members = append(members, user)
	}
	return members, rows.Err()
}

func (r *TeamRepository) AddMember(teamName string, userID int) error {
	_, err := r.db.Exec(
		"INSERT INTO team_members (team_name, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		teamName, userID,
	)
	return err
}

func (r *TeamRepository) RemoveMember(teamName string, userID int) error {
	_, err := r.db.Exec(
		"DELETE FROM team_members WHERE team_name = $1 AND user_id = $2",
		teamName, userID,
	)
	return err
}

func (r *TeamRepository) GetUserTeam(userID int) (string, error) {
	var teamName string
	err := r.db.QueryRow(
		"SELECT team_name FROM team_members WHERE user_id = $1 LIMIT 1",
		userID,
	).Scan(&teamName)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return teamName, err
}

