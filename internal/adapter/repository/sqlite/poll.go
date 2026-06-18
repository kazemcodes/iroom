package repository

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/iroom/iroom/internal/domain/entity"
)

type PollRepo struct {
	db *sql.DB
}

func NewPollRepo(db *sql.DB) *PollRepo {
	return &PollRepo{db: db}
}

func (r *PollRepo) Create(p *entity.Poll) error {
	optionsJSON, err := json.Marshal(p.Options)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(
		`INSERT INTO polls (session_id, question, options, is_active) VALUES (?, ?, ?, ?)`,
		p.SessionID, p.Question, string(optionsJSON), p.IsActive,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

func (r *PollRepo) GetByID(id int64) (*entity.Poll, error) {
	p := &entity.Poll{}
	var optionsJSON string
	err := r.db.QueryRow(
		`SELECT id, session_id, question, options, is_active, created_at FROM polls WHERE id = ?`, id,
	).Scan(&p.ID, &p.SessionID, &p.Question, &optionsJSON, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(optionsJSON), &p.Options); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PollRepo) ListBySession(sessionID int64) ([]*entity.Poll, error) {
	rows, err := r.db.Query(
		`SELECT id, session_id, question, options, is_active, created_at 
		 FROM polls 
		 WHERE session_id = ? 
		 ORDER BY created_at DESC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var polls []*entity.Poll
	for rows.Next() {
		p := &entity.Poll{}
		var optionsJSON string
		if err := rows.Scan(&p.ID, &p.SessionID, &p.Question, &optionsJSON, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(optionsJSON), &p.Options); err != nil {
			return nil, err
		}
		polls = append(polls, p)
	}
	return polls, nil
}

func (r *PollRepo) Close(id int64) error {
	_, err := r.db.Exec(
		`UPDATE polls SET is_active = FALSE WHERE id = ?`,
		id,
	)
	return err
}

func (r *PollRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM polls WHERE id = ?`, id)
	return err
}

func (r *PollRepo) Vote(vote *entity.PollVote) error {
	// Use INSERT OR REPLACE to handle re-voting (update existing vote)
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO poll_votes (poll_id, user_id, option_index) VALUES (?, ?, ?)`,
		vote.PollID, vote.UserID, vote.OptionIndex,
	)
	return err
}

func (r *PollRepo) GetVote(pollID, userID int64) (*entity.PollVote, error) {
	v := &entity.PollVote{}
	err := r.db.QueryRow(
		`SELECT id, poll_id, user_id, option_index, created_at FROM poll_votes WHERE poll_id = ? AND user_id = ?`,
		pollID, userID,
	).Scan(&v.ID, &v.PollID, &v.UserID, &v.OptionIndex, &v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (r *PollRepo) GetResults(pollID int64) (*entity.PollResults, error) {
	poll, err := r.GetByID(pollID)
	if err != nil {
		return nil, err
	}

	options := make([]string, 0)
	if poll.Options != "" {
		for _, o := range strings.Split(poll.Options, ",") {
			options = append(options, strings.TrimSpace(o))
		}
	}

	results := &entity.PollResults{
		PollID:   poll.ID,
		Question: poll.Question,
		Options:  options,
		Votes:    make(map[int]int),
	}

	// Get vote counts per option
	rows, err := r.db.Query(
		`SELECT option_index, COUNT(*) as count 
		 FROM poll_votes 
		 WHERE poll_id = ? 
		 GROUP BY option_index`,
		pollID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var optionIndex, count int
		if err := rows.Scan(&optionIndex, &count); err != nil {
			return nil, err
		}
		if optionIndex >= 0 && optionIndex < len(results.Votes) {
			results.Votes[optionIndex] = count
			results.TotalVotes += count
		}
	}

	return results, nil
}

func (r *PollRepo) HasUserVoted(pollID, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM poll_votes WHERE poll_id = ? AND user_id = ?)`,
		pollID, userID,
	).Scan(&exists)
	return exists, err
}
