package sop

import (
	"database/sql"
)

type SOPRow struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Nick        string `json:"nick,omitempty"`
	Description string `json:"description,omitempty"`
	Content     string `json:"content,omitempty"`
	Category    string `json:"category,omitempty"`
	Tags        string `json:"tags,omitempty"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type QueryResult struct {
	Count int
	Rows  []SOPRow
}

func GetSOPList() (*QueryResult, error) {
	rows, err := pgDB.Query(`
		SELECT uuid, name, nick, description, content, category, tags, created_at, updated_at
		FROM sops
		WHERE is_deleted = FALSE
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SOPRow
	for rows.Next() {
		var r SOPRow
		var nick, description, content, category, tags sql.NullString
		var createdAt, updatedAt sql.NullString

		err := rows.Scan(&r.UUID, &r.Name, &nick, &description, &content, &category, &tags, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		r.Nick = nick.String
		r.Description = description.String
		r.Content = content.String
		r.Category = category.String
		r.Tags = tags.String
		r.CreatedAt = createdAt.String
		r.UpdatedAt = updatedAt.String

		results = append(results, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if results == nil {
		results = []SOPRow{}
	}

	return &QueryResult{
		Count: len(results),
		Rows:  results,
	}, nil
}

func GetSOPByUUID(sopUUID string) (*QueryResult, error) {
	row := pgDB.QueryRow(`
		SELECT uuid, name, nick, description, content, category, tags, created_at, updated_at
		FROM sops
		WHERE uuid = $1 AND is_deleted = FALSE
	`, sopUUID)

	var r SOPRow
	var nick, description, content, category, tags sql.NullString
	var createdAt, updatedAt sql.NullString

	err := row.Scan(&r.UUID, &r.Name, &nick, &description, &content, &category, &tags, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	r.Nick = nick.String
	r.Description = description.String
	r.Content = content.String
	r.Category = category.String
	r.Tags = tags.String
	r.CreatedAt = createdAt.String
	r.UpdatedAt = updatedAt.String

	return &QueryResult{
		Count: 1,
		Rows:  []SOPRow{r},
	}, nil
}
