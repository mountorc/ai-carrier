package sop

import (
	"database/sql"
	"encoding/json"
	"time"
)

type SOPRow struct {
	UUID        string          `json:"uuid"`
	Name        string          `json:"name"`
	Nick        string          `json:"nick,omitempty"`
	Description string          `json:"description,omitempty"`
	Content     string          `json:"content,omitempty"`
	Config      json.RawMessage `json:"config,omitempty"`
	Category    string          `json:"category,omitempty"`
	Tags        string          `json:"tags,omitempty"`
	Status      string          `json:"status,omitempty"`
	Version     string          `json:"version,omitempty"`
	Project     string          `json:"project,omitempty"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

type QueryResult struct {
	Count int
	Rows  []SOPRow
}

func GetSOPList() (*QueryResult, error) {
	rows, err := pgDB.Query(`
		SELECT uuid, name, nick, description, content, config, category, tags, status, version, project, created_at, updated_at
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
		var nick, description, content, category, tags, status, version, project sql.NullString
		var configData []byte
		var createdAt, updatedAt sql.NullString

		err := rows.Scan(&r.UUID, &r.Name, &nick, &description, &content, &configData, &category, &tags, &status, &version, &project, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}

		r.Nick = nick.String
		r.Description = description.String
		r.Content = content.String
		r.Category = category.String
		r.Tags = tags.String
		r.Status = status.String
		r.Version = version.String
		r.Project = project.String
		r.CreatedAt = createdAt.String
		r.UpdatedAt = updatedAt.String
		if configData != nil {
			r.Config = json.RawMessage(configData)
		}

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
		SELECT uuid, name, nick, description, content, config, category, tags, status, version, project, created_at, updated_at
		FROM sops
		WHERE uuid = $1 AND is_deleted = FALSE
	`, sopUUID)

	var r SOPRow
	var nick, description, content, category, tags, status, version, project sql.NullString
	var configData []byte
	var createdAt, updatedAt sql.NullString

	err := row.Scan(&r.UUID, &r.Name, &nick, &description, &content, &configData, &category, &tags, &status, &version, &project, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	r.Nick = nick.String
	r.Description = description.String
	r.Content = content.String
	r.Category = category.String
	r.Tags = tags.String
	r.Status = status.String
	r.Version = version.String
	r.Project = project.String
	r.CreatedAt = createdAt.String
	r.UpdatedAt = updatedAt.String
	if configData != nil {
		r.Config = json.RawMessage(configData)
	}

	return &QueryResult{
		Count: 1,
		Rows:  []SOPRow{r},
	}, nil
}

func CreateSOP(name, nick, description, content, category, tags, status, version, project string, config json.RawMessage) (*SOPRow, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	if status == "" {
		status = "draft"
	}
	if version == "" {
		version = "1.0"
	}

	var configBytes []byte
	if config != nil {
		configBytes = config
	} else {
		configBytes = []byte("{}")
	}

	row := pgDB.QueryRow(`
		INSERT INTO sops (uuid, name, nick, description, content, config, category, tags, status, version, project, created_at, updated_at, is_deleted)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9, $10, $11, $12, FALSE)
		RETURNING uuid, name, nick, description, content, config, category, tags, status, version, project, created_at, updated_at
	`, name, nullStr(nick), nullStr(description), nullStr(content), configBytes, nullStr(category), nullStr(tags), status, version, nullStr(project), now, now)

	var r SOPRow
	var n, d, c, cat, t, s, v, p sql.NullString
	var cfg []byte
	var ca, ua sql.NullString

	err := row.Scan(&r.UUID, &r.Name, &n, &d, &c, &cfg, &cat, &t, &s, &v, &p, &ca, &ua)
	if err != nil {
		return nil, err
	}

	r.Nick = n.String
	r.Description = d.String
	r.Content = c.String
	r.Category = cat.String
	r.Tags = t.String
	r.Status = s.String
	r.Version = v.String
	r.Project = p.String
	r.CreatedAt = ca.String
	r.UpdatedAt = ua.String
	if cfg != nil {
		r.Config = json.RawMessage(cfg)
	}

	return &r, nil
}

func UpdateSOP(uuid, name, nick, description, content, category, tags, status, version, project string, config json.RawMessage) (*SOPRow, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	var configBytes []byte
	if config != nil {
		configBytes = config
	}

	row := pgDB.QueryRow(`
		UPDATE sops SET
			name = COALESCE(NULLIF($2, ''), name),
			nick = $3,
			description = $4,
			content = $5,
			config = COALESCE($6::jsonb, config),
			category = $7,
			tags = $8,
			status = COALESCE(NULLIF($9, ''), status),
			version = COALESCE(NULLIF($10, ''), version),
			project = $11,
			updated_at = $12
		WHERE uuid = $1 AND is_deleted = FALSE
		RETURNING uuid, name, nick, description, content, config, category, tags, status, version, project, created_at, updated_at
	`, uuid, name, nullStr(nick), nullStr(description), nullStr(content), configBytes, nullStr(category), nullStr(tags), status, version, nullStr(project), now)

	var r SOPRow
	var n, d, c, cat, t, s, v, p sql.NullString
	var cfg []byte
	var ca, ua sql.NullString

	err := row.Scan(&r.UUID, &r.Name, &n, &d, &c, &cfg, &cat, &t, &s, &v, &p, &ca, &ua)
	if err != nil {
		return nil, err
	}

	r.Nick = n.String
	r.Description = d.String
	r.Content = c.String
	r.Category = cat.String
	r.Tags = t.String
	r.Status = s.String
	r.Version = v.String
	r.Project = p.String
	r.CreatedAt = ca.String
	r.UpdatedAt = ua.String
	if cfg != nil {
		r.Config = json.RawMessage(cfg)
	}

	return &r, nil
}

func DeleteSOP(uuid string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := pgDB.Exec(`
		UPDATE sops SET is_deleted = TRUE, updated_at = $2
		WHERE uuid = $1 AND is_deleted = FALSE
	`, uuid, now)
	return err
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
