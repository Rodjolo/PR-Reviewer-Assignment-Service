package repository

import (
	"database/sql"
	"pr-reviewer-service/internal/models"
	"time"
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
	// Используем JOIN для загрузки всех данных за один запрос
	rows, err := r.db.Query(`
		SELECT DISTINCT pr.id, pr.title, pr.author_id, pr.status, prr.reviewer_id
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.id = prr.pr_id
		WHERE pr.author_id = $1 OR prr.reviewer_id = $1
		ORDER BY pr.id, prr.reviewer_id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prsMap := make(map[int]*models.PR)
	for rows.Next() {
		var prID, authorID int
		var title, status string
		var reviewerID sql.NullInt64

		if err := rows.Scan(&prID, &title, &authorID, &status, &reviewerID); err != nil {
			return nil, err
		}

		// Создаем PR, если его еще нет
		if _, exists := prsMap[prID]; !exists {
			prsMap[prID] = &models.PR{
				ID:        prID,
				Title:     title,
				AuthorID:  authorID,
				Status:    models.PRStatus(status),
				Reviewers: []int{},
			}
		}

		// Добавляем ревьювера, если он есть
		if reviewerID.Valid {
			prsMap[prID].Reviewers = append(prsMap[prID].Reviewers, int(reviewerID.Int64))
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Преобразуем map в slice
	prs := make([]models.PR, 0, len(prsMap))
	for _, pr := range prsMap {
		prs = append(prs, *pr)
	}

	return prs, nil
}

func (r *PRRepository) GetAll() ([]models.PR, error) {
	// Используем JOIN для загрузки всех данных за один запрос
	rows, err := r.db.Query(`
		SELECT pr.id, pr.title, pr.author_id, pr.status, prr.reviewer_id
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.id = prr.pr_id
		ORDER BY pr.id, prr.reviewer_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prsMap := make(map[int]*models.PR)
	for rows.Next() {
		var prID, authorID int
		var title, status string
		var reviewerID sql.NullInt64

		if err := rows.Scan(&prID, &title, &authorID, &status, &reviewerID); err != nil {
			return nil, err
		}

		// Создаем PR, если его еще нет
		if _, exists := prsMap[prID]; !exists {
			prsMap[prID] = &models.PR{
				ID:        prID,
				Title:     title,
				AuthorID:  authorID,
				Status:    models.PRStatus(status),
				Reviewers: []int{},
			}
		}

		// Добавляем ревьювера, если он есть
		if reviewerID.Valid {
			prsMap[prID].Reviewers = append(prsMap[prID].Reviewers, int(reviewerID.Int64))
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Преобразуем map в slice
	prs := make([]models.PR, 0, len(prsMap))
	for _, pr := range prsMap {
		prs = append(prs, *pr)
	}

	return prs, nil
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
	stats := make(map[string]int)
	
	err := r.db.QueryRow(`
		SELECT 
			(SELECT COUNT(*) FROM users) as total_users,
			(SELECT COUNT(*) FROM users WHERE is_active = true) as active_users,
			(SELECT COUNT(*) FROM teams) as total_teams,
			(SELECT COUNT(*) FROM pull_requests) as total_prs,
			(SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN') as open_prs,
			(SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED') as merged_prs
	`).Scan(
		&stats["total_users"],
		&stats["active_users"],
		&stats["total_teams"],
		&stats["total_prs"],
		&stats["open_prs"],
		&stats["merged_prs"],
	)
	
	if err != nil {
		return nil, err
	}

	return stats, nil
}

