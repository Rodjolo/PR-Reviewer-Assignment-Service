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
	rows, err := r.db.Query(`
		SELECT DISTINCT pr.id, pr.title, pr.author_id, pr.status
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.id = prr.pr_id
		WHERE pr.author_id = $1 OR prr.reviewer_id = $1
		ORDER BY pr.id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PR
	for rows.Next() {
		var pr models.PR
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	// Загружаем ревьюверов для каждого PR
	for i := range prs {
		reviewers, err := r.getReviewers(prs[i].ID)
		if err != nil {
			return nil, err
		}
		prs[i].Reviewers = reviewers
	}

	return prs, rows.Err()
}

func (r *PRRepository) GetAll() ([]models.PR, error) {
	rows, err := r.db.Query("SELECT id, title, author_id, status FROM pull_requests ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []models.PR
	for rows.Next() {
		var pr models.PR
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	// Загружаем ревьюверов для каждого PR
	for i := range prs {
		reviewers, err := r.getReviewers(prs[i].ID)
		if err != nil {
			return nil, err
		}
		prs[i].Reviewers = reviewers
	}

	return prs, rows.Err()
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
	stats := make(map[string]int)
	var count int

	// Общее количество пользователей
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["total_users"] = count

	// Активные пользователи
	err = r.db.QueryRow("SELECT COUNT(*) FROM users WHERE is_active = true").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["active_users"] = count

	// Общее количество команд
	err = r.db.QueryRow("SELECT COUNT(*) FROM teams").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["total_teams"] = count

	// Общее количество PR
	err = r.db.QueryRow("SELECT COUNT(*) FROM pull_requests").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["total_prs"] = count

	// Открытые PR
	err = r.db.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE status = 'OPEN'").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["open_prs"] = count

	// Мерженные PR
	err = r.db.QueryRow("SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'").Scan(&count)
	if err != nil {
		return nil, err
	}
	stats["merged_prs"] = count

	return stats, nil
}

