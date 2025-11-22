package repository

import (
	"database/sql"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"

	"github.com/lib/pq"
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
	// Оптимизированный подход: сначала получаем команды, потом участников
	// Это быстрее при большом количестве данных
	teamRows, err := r.db.Query("SELECT name FROM teams ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer teamRows.Close()

	var teams []models.Team
	for teamRows.Next() {
		var team models.Team
		if err := teamRows.Scan(&team.Name); err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}

	if err := teamRows.Err(); err != nil {
		return nil, err
	}

	// Загружаем участников для всех команд одним запросом
	if len(teams) == 0 {
		return teams, nil
	}

	teamNames := make([]string, len(teams))
	for i, team := range teams {
		teamNames[i] = team.Name
	}

	memberRows, err := r.db.Query(`
		SELECT tm.team_name, u.id, u.name, u.is_active
		FROM team_members tm
		INNER JOIN users u ON tm.user_id = u.id
		WHERE tm.team_name = ANY($1::text[])
		ORDER BY tm.team_name, u.id
	`, pq.Array(teamNames))
	if err != nil {
		return nil, err
	}
	defer memberRows.Close()

	// Создаем map для быстрого доступа
	teamsMap := make(map[string]*models.Team)
	for i := range teams {
		teams[i].Members = []models.User{}
		teamsMap[teams[i].Name] = &teams[i]
	}

	// Заполняем участников
	for memberRows.Next() {
		var teamName string
		var user models.User
		if err := memberRows.Scan(&teamName, &user.ID, &user.Name, &user.IsActive); err != nil {
			return nil, err
		}
		if team, exists := teamsMap[teamName]; exists {
			team.Members = append(team.Members, user)
		}
	}

	return teams, memberRows.Err()
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
