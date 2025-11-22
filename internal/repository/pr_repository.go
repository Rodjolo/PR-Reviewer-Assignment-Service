package repository

import (
	"database/sql"
	"github.com/Rodjolo/pr-reviewer-service/pkg/models"
	"math/rand"
	"time"

	"github.com/lib/pq"
)

type PRRepository struct {
	db *sql.DB
}

func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{db: db}
}

func (r *PRRepository) Create(pr *models.PR) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = tx.QueryRow(
		"INSERT INTO pull_requests (title, author_id, status) VALUES ($1, $2, $3) RETURNING id",
		pr.Title, pr.AuthorID, pr.Status,
	).Scan(&pr.ID)
	if err != nil {
		return err
	}

	// Добавляем ревьюверов
	for _, reviewerID := range pr.Reviewers {
		_, err = tx.Exec(
			"INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)",
			pr.ID, reviewerID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PRRepository) GetByID(id int) (*models.PR, error) {
	pr := &models.PR{}
	err := r.db.QueryRow(
		"SELECT id, title, author_id, status FROM pull_requests WHERE id = $1",
		id,
	).Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Загружаем ревьюверов
	rows, err := r.db.Query(
		"SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1 ORDER BY reviewer_id",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []int
	for rows.Next() {
		var reviewerID int
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}
	pr.Reviewers = reviewers

	return pr, rows.Err()
}

func (r *PRRepository) GetByUserID(userID int) ([]models.PR, error) {
	// Оптимизированный подход: сначала получаем PR, потом ревьюверов
	prRows, err := r.db.Query(`
		SELECT DISTINCT pr.id, pr.title, pr.author_id, pr.status
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.id = prr.pr_id
		WHERE pr.author_id = $1 OR prr.reviewer_id = $1
		ORDER BY pr.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer prRows.Close()

	var prs []models.PR
	for prRows.Next() {
		var pr models.PR
		if err := prRows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		pr.Reviewers = []int{}
		prs = append(prs, pr)
	}

	if err := prRows.Err(); err != nil {
		return nil, err
	}

	// Загружаем ревьюверов для всех PR одним запросом
	if len(prs) == 0 {
		return prs, nil
	}

	prIDs := make([]int, len(prs))
	for i, pr := range prs {
		prIDs[i] = pr.ID
	}

	reviewerRows, err := r.db.Query(`
		SELECT pr_id, reviewer_id
		FROM pr_reviewers
		WHERE pr_id = ANY($1::int[])
		ORDER BY pr_id, reviewer_id
	`, pq.Array(prIDs))
	if err != nil {
		return nil, err
	}
	defer reviewerRows.Close()

	// Создаем map для быстрого доступа
	prsMap := make(map[int]*models.PR)
	for i := range prs {
		prsMap[prs[i].ID] = &prs[i]
	}

	// Заполняем ревьюверов
	for reviewerRows.Next() {
		var prID, reviewerID int
		if err := reviewerRows.Scan(&prID, &reviewerID); err != nil {
			return nil, err
		}
		if pr, exists := prsMap[prID]; exists {
			pr.Reviewers = append(pr.Reviewers, reviewerID)
		}
	}

	return prs, reviewerRows.Err()
}

func (r *PRRepository) GetAll() ([]models.PR, error) {
	// Оптимизированный подход: сначала получаем PR, потом ревьюверов
	// Это быстрее при большом количестве данных
	prRows, err := r.db.Query("SELECT id, title, author_id, status FROM pull_requests ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer prRows.Close()

	var prs []models.PR
	for prRows.Next() {
		var pr models.PR
		if err := prRows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		pr.Reviewers = []int{}
		prs = append(prs, pr)
	}

	if err := prRows.Err(); err != nil {
		return nil, err
	}

	// Загружаем ревьюверов для всех PR одним запросом
	if len(prs) == 0 {
		return prs, nil
	}

	prIDs := make([]int, len(prs))
	for i, pr := range prs {
		prIDs[i] = pr.ID
	}

	reviewerRows, err := r.db.Query(`
		SELECT pr_id, reviewer_id
		FROM pr_reviewers
		WHERE pr_id = ANY($1::int[])
		ORDER BY pr_id, reviewer_id
	`, pq.Array(prIDs))
	if err != nil {
		return nil, err
	}
	defer reviewerRows.Close()

	// Создаем map для быстрого доступа
	prsMap := make(map[int]*models.PR)
	for i := range prs {
		prsMap[prs[i].ID] = &prs[i]
	}

	// Заполняем ревьюверов
	for reviewerRows.Next() {
		var prID, reviewerID int
		if err := reviewerRows.Scan(&prID, &reviewerID); err != nil {
			return nil, err
		}
		if pr, exists := prsMap[prID]; exists {
			pr.Reviewers = append(pr.Reviewers, reviewerID)
		}
	}

	return prs, reviewerRows.Err()
}

func (r *PRRepository) getReviewers(prID int) ([]int, error) {
	rows, err := r.db.Query(
		"SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1 ORDER BY reviewer_id",
		prID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []int
	for rows.Next() {
		var reviewerID int
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}
	return reviewers, rows.Err()
}

func (r *PRRepository) UpdateStatus(id int, status models.PRStatus) error {
	now := time.Now()
	_, err := r.db.Exec(
		"UPDATE pull_requests SET status = $1, merged_at = $2 WHERE id = $3",
		status, now, id,
	)
	return err
}

func (r *PRRepository) ReassignReviewer(prID int, oldReviewerID int, newReviewerID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Удаляем старого ревьювера
	_, err = tx.Exec(
		"DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2",
		prID, oldReviewerID,
	)
	if err != nil {
		return err
	}

	// Добавляем нового ревьювера
	_, err = tx.Exec(
		"INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)",
		prID, newReviewerID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PRRepository) GetStats() (map[string]int, error) {
	// Используем один запрос с подзапросами для получения всей статистики
	stats := make(map[string]int, 6)

	var (
		totalUsers  int
		activeUsers int
		totalTeams  int
		totalPRs    int
		openPRs     int
		mergedPRs   int
	)

	err := r.db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM users WHERE is_active = true) as active_users,
			(SELECT COUNT(*) FROM teams) as total_teams,
			(SELECT COUNT(*) FROM pull_requests) as total_prs,
			(SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN') as open_prs,
			(SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED') as merged_prs
	`).Scan(
		&totalUsers,
		&activeUsers,
		&totalTeams,
		&totalPRs,
		&openPRs,
		&mergedPRs,
	)

	if err != nil {
		return nil, err
	}

	stats["total_users"] = totalUsers
	stats["active_users"] = activeUsers
	stats["total_teams"] = totalTeams
	stats["total_prs"] = totalPRs
	stats["open_prs"] = openPRs
	stats["merged_prs"] = mergedPRs

	return stats, nil
}

// GetOpenPRsWithReviewers получает открытые PR с ревьюверами из указанного списка пользователей
// Возвращает map[prID][]reviewerID для эффективной обработки
func (r *PRRepository) GetOpenPRsWithReviewers(userIDs []int) (map[int][]int, error) {
	if len(userIDs) == 0 {
		return make(map[int][]int), nil
	}

	// Используем ANY для эффективного поиска в массиве
	query := `
		SELECT pr.id, prr.reviewer_id
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.id = prr.pr_id
		WHERE pr.status = $1 AND prr.reviewer_id = ANY($2::int[])
		ORDER BY pr.id, prr.reviewer_id
	`

	rows, err := r.db.Query(query, models.PRStatusOpen, pq.Array(userIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]int)
	for rows.Next() {
		var prID, reviewerID int
		if err := rows.Scan(&prID, &reviewerID); err != nil {
			return nil, err
		}
		result[prID] = append(result[prID], reviewerID)
	}

	return result, rows.Err()
}

// BulkReassignReviewers безопасно переназначает ревьюверов в открытых PR
// prReviewerMap - map[prID][]reviewerID - открытые PR с ревьюверами, которых нужно заменить
// teamName - имя команды для поиска новых ревьюверов
// excludeUserIDs - список ID пользователей, которых нужно исключить из кандидатов
// Возвращает количество переназначений
func (r *PRRepository) BulkReassignReviewers(prReviewerMap map[int][]int, teamName string, excludeUserIDs []int) (int, error) {
	if len(prReviewerMap) == 0 {
		return 0, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Получаем активных пользователей команды для переназначения
	var candidates []int
	var candidatesRows *sql.Rows
	if len(excludeUserIDs) > 0 {
		candidatesRows, err = tx.Query(`
			SELECT u.id
			FROM users u
			INNER JOIN team_members tm ON u.id = tm.user_id
			WHERE tm.team_name = $1 AND u.is_active = true AND u.id != ALL($2::int[])
			ORDER BY RANDOM()
			LIMIT 100
		`, teamName, pq.Array(excludeUserIDs))
	} else {
		candidatesRows, err = tx.Query(`
			SELECT u.id
			FROM users u
			INNER JOIN team_members tm ON u.id = tm.user_id
			WHERE tm.team_name = $1 AND u.is_active = true
			ORDER BY RANDOM()
			LIMIT 100
		`, teamName)
	}

	if err != nil {
		return 0, err
	}
	defer candidatesRows.Close()

	for candidatesRows.Next() {
		var candidateID int
		if err := candidatesRows.Scan(&candidateID); err != nil {
			return 0, err
		}
		candidates = append(candidates, candidateID)
	}

	if len(candidates) == 0 {
		// Нет доступных кандидатов - просто удаляем ревьюверов
		reassignments := 0
		for prID, oldReviewerIDs := range prReviewerMap {
			for _, oldReviewerID := range oldReviewerIDs {
				_, err = tx.Exec(
					"DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2",
					prID, oldReviewerID,
				)
				if err != nil {
					return 0, err
				}
				reassignments++
			}
		}
		return reassignments, tx.Commit()
	}

	// Получаем авторов всех PR одним запросом для оптимизации
	prIDs := make([]int, 0, len(prReviewerMap))
	for prID := range prReviewerMap {
		prIDs = append(prIDs, prID)
	}

	authorsMap := make(map[int]int)
	if len(prIDs) > 0 {
		rows, err := tx.Query(`
			SELECT id, author_id FROM pull_requests WHERE id = ANY($1::int[])
		`, pq.Array(prIDs))
		if err != nil {
			return 0, err
		}
		for rows.Next() {
			var prID, authorID int
			if err := rows.Scan(&prID, &authorID); err != nil {
				rows.Close()
				return 0, err
			}
			authorsMap[prID] = authorID
		}
		rows.Close()
	}

	// Переназначаем ревьюверов
	reassignments := 0
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for prID, oldReviewerIDs := range prReviewerMap {
		authorID := authorsMap[prID]

		// Фильтруем кандидатов, исключая автора
		availableCandidates := make([]int, 0)
		for _, candidateID := range candidates {
			if candidateID != authorID {
				availableCandidates = append(availableCandidates, candidateID)
			}
		}

		if len(availableCandidates) == 0 {
			// Нет доступных кандидатов - удаляем ревьюверов
			for _, oldReviewerID := range oldReviewerIDs {
				_, err = tx.Exec(
					"DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2",
					prID, oldReviewerID,
				)
				if err != nil {
					return 0, err
				}
				reassignments++
			}
			continue
		}

		// Переназначаем каждого старого ревьювера
		for _, oldReviewerID := range oldReviewerIDs {
			// Выбираем случайного нового ревьювера
			newReviewerID := availableCandidates[rng.Intn(len(availableCandidates))]

			// Удаляем старого ревьювера
			_, err = tx.Exec(
				"DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2",
				prID, oldReviewerID,
			)
			if err != nil {
				return 0, err
			}

			// Добавляем нового ревьювера (если его еще нет)
			_, err = tx.Exec(
				"INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
				prID, newReviewerID,
			)
			if err != nil {
				return 0, err
			}
			reassignments++
		}
	}

	return reassignments, tx.Commit()
}
