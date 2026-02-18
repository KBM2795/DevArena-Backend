package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ChallengeQueryParams holds the query parameters for fetching challenges
type ChallengeQueryParams struct {
	Page       int
	Limit      int
	Difficulty string
	Type       string
	Search     string
	Sort       string
	Tags       []string
}

// ChallengeResponse represents a challenge in API responses
type ChallengeResponse struct {
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	DescriptionMD   string          `json:"description_md"`
	Difficulty      string          `json:"difficulty"`
	Type            string          `json:"type"`
	MaxScore        int             `json:"max_score"`
	RepoTemplateURL string          `json:"repo_template_url,omitempty"`
	Requirements    []string        `json:"requirements"`
	TechStack       []string        `json:"tech_stack"`
	EstimatedHours  int             `json:"estimated_hours"`
	Rubric          json.RawMessage `json:"rubric,omitempty"`
	IsPublished     bool            `json:"is_published"`
	Tags            []string        `json:"tags"`
	SuccessRate     float64         `json:"success_rate"`
	CreatedAt       string          `json:"created_at"`
}

// ChallengeListResponse is the paginated response for challenges
type ChallengeListResponse struct {
	Data       []ChallengeResponse `json:"data"`
	Pagination PaginationMeta      `json:"pagination"`
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// TagResponse represents a tag in API responses
type TagResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Category string `json:"category"`
	Color    string `json:"color"`
	Count    int    `json:"count"` // Number of challenges with this tag
}

// TagCategoryResponse groups tags by category
type TagCategoryResponse struct {
	Category string        `json:"category"`
	Tags     []TagResponse `json:"tags"`
}

// GetTags retrieves all available tags grouped by category with challenge counts
func (db *Database) GetTags() ([]TagCategoryResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT 
			t.id,
			t.name,
			t.slug,
			COALESCE(t.category, 'other') as category,
			COALESCE(t.color, '#666666') as color,
			COUNT(ct.challenge_id) as challenge_count
		FROM tags t
		LEFT JOIN challenge_tags ct ON ct.tag_id = t.id
		LEFT JOIN challenges c ON c.id = ct.challenge_id AND c.is_published = true
		GROUP BY t.id, t.name, t.slug, t.category, t.color
		HAVING COUNT(ct.challenge_id) > 0
		ORDER BY t.category, challenge_count DESC, t.name
	`

	rows, err := db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	// Group tags by category
	categoryMap := make(map[string][]TagResponse)
	for rows.Next() {
		var tag TagResponse
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.Category, &tag.Color, &tag.Count)
		if err != nil {
			continue
		}
		categoryMap[tag.Category] = append(categoryMap[tag.Category], tag)
	}

	// Convert map to slice with proper order
	categoryOrder := []string{"frontend", "backend", "database", "ai", "fundamentals", "other"}
	result := []TagCategoryResponse{}

	for _, cat := range categoryOrder {
		if tags, ok := categoryMap[cat]; ok {
			result = append(result, TagCategoryResponse{
				Category: cat,
				Tags:     tags,
			})
			delete(categoryMap, cat)
		}
	}

	// Add any remaining categories
	for cat, tags := range categoryMap {
		result = append(result, TagCategoryResponse{
			Category: cat,
			Tags:     tags,
		})
	}

	return result, nil
}

// GetChallenges retrieves challenges with pagination and filters
func (db *Database) GetChallenges(params ChallengeQueryParams) (*ChallengeListResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}

	// Build WHERE clause
	whereClauses := []string{"c.is_published = true"}
	args := []interface{}{}
	argIndex := 1

	if params.Difficulty != "" && params.Difficulty != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf("LOWER(c.difficulty) = LOWER($%d)", argIndex))
		args = append(args, params.Difficulty)
		argIndex++
	}

	// Filter by tags (supports multiple, uses OR logic)
	if len(params.Tags) > 0 {
		tagConditions := []string{}
		for _, tag := range params.Tags {
			tagConditions = append(tagConditions, fmt.Sprintf(`
				EXISTS (
					SELECT 1 FROM challenge_tags ct 
					JOIN tags t ON t.id = ct.tag_id 
					WHERE ct.challenge_id = c.id AND LOWER(t.name) = LOWER($%d)
				)
				OR c.tech_stack::text ILIKE $%d
			`, argIndex, argIndex+1))
			args = append(args, tag, "%"+tag+"%")
			argIndex += 2
		}
		// Combine with OR - match any of the selected tags
		whereClauses = append(whereClauses, "("+strings.Join(tagConditions, " OR ")+")")
	}

	if params.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(c.title) LIKE LOWER($%d) OR LOWER(c.description) LIKE LOWER($%d))", argIndex, argIndex))
		args = append(args, "%"+params.Search+"%")
		argIndex++
	}

	whereClause := strings.Join(whereClauses, " AND ")

	// Build ORDER BY clause
	orderBy := "c.created_at ASC" // default: oldest
	switch params.Sort {
	case "popular":
		orderBy = "c.max_score DESC"
	case "difficulty":
		orderBy = "CASE c.difficulty WHEN 'Easy' THEN 1 WHEN 'Medium' THEN 2 WHEN 'Hard' THEN 3 END ASC"
	case "oldest":
		orderBy = "c.created_at ASC"
	}

	// Get total count
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM challenges c WHERE %s`, whereClause)
	var total int
	err := db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count challenges: %w", err)
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.Limit
	totalPages := (total + params.Limit - 1) / params.Limit

	// Get challenges with tags
	query := fmt.Sprintf(`
		SELECT 
			c.id,
			c.title,
			c.description,
			COALESCE(c.description_md, ''),
			c.difficulty,
			c.type,
			c.max_score,
			COALESCE(c.repo_template_url, ''),
			c.requirements,
			c.tech_stack,
			c.estimated_hours,
			c.is_published,
			c.created_at,
			COALESCE(c.rubric, '{}'::jsonb),
			COALESCE(
				(SELECT array_agg(t.name) FROM tags t 
				 JOIN challenge_tags ct ON ct.tag_id = t.id 
				 WHERE ct.challenge_id = c.id),
				ARRAY[]::text[]
			) as tags
		FROM challenges c
		WHERE %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argIndex, argIndex+1)

	args = append(args, params.Limit, offset)

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query challenges: %w", err)
	}
	defer rows.Close()

	challenges := []ChallengeResponse{}
	for rows.Next() {
		var c ChallengeResponse
		var requirementsJSON, techStackJSON string
		var tags []string
		var createdAt time.Time
		var rubricBytes []byte

		err := rows.Scan(
			&c.ID,
			&c.Title,
			&c.Description,
			&c.DescriptionMD,
			&c.Difficulty,
			&c.Type,
			&c.MaxScore,
			&c.RepoTemplateURL,
			&requirementsJSON,
			&techStackJSON,
			&c.EstimatedHours,
			&c.IsPublished,
			&createdAt,
			&rubricBytes,
			&tags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan challenge: %w", err)
		}

		// Parse JSON arrays
		json.Unmarshal([]byte(requirementsJSON), &c.Requirements)
		json.Unmarshal([]byte(techStackJSON), &c.TechStack)

		c.Rubric = rubricBytes
		c.Tags = tags
		if c.Tags == nil {
			c.Tags = []string{}
		}
		c.CreatedAt = createdAt.Format(time.RFC3339)
		c.SuccessRate = 0 // TODO: Calculate from submissions

		challenges = append(challenges, c)
	}

	return &ChallengeListResponse{
		Data: challenges,
		Pagination: PaginationMeta{
			Page:       params.Page,
			Limit:      params.Limit,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    params.Page < totalPages,
			HasPrev:    params.Page > 1,
		},
	}, nil
}

// GetChallengeByID retrieves a single challenge by ID
func (db *Database) GetChallengeByID(id string) (*ChallengeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT 
			c.id,
			c.title,
			c.description,
			COALESCE(c.description_md, ''),
			c.difficulty,
			c.type,
			c.max_score,
			COALESCE(c.repo_template_url, ''),
			c.requirements,
			c.tech_stack,
			c.estimated_hours,
			c.is_published,
			c.created_at,
			COALESCE(c.rubric, '{}'::jsonb),
			COALESCE(
				(SELECT array_agg(t.name) FROM tags t 
				 JOIN challenge_tags ct ON ct.tag_id = t.id 
				 WHERE ct.challenge_id = c.id),
				ARRAY[]::text[]
			) as tags
		FROM challenges c
		WHERE c.id = $1 AND c.is_published = true
	`

	var c ChallengeResponse
	var requirementsJSON, techStackJSON string
	var tags []string
	var createdAt time.Time
	var rubricBytes []byte

	err := db.Pool.QueryRow(ctx, query, id).Scan(
		&c.ID,
		&c.Title,
		&c.Description,
		&c.DescriptionMD,
		&c.Difficulty,
		&c.Type,
		&c.MaxScore,
		&c.RepoTemplateURL,
		&requirementsJSON,
		&techStackJSON,
		&c.EstimatedHours,
		&c.IsPublished,
		&createdAt,
		&rubricBytes,
		&tags,
	)
	if err != nil {
		return nil, fmt.Errorf("challenge not found: %w", err)
	}

	json.Unmarshal([]byte(requirementsJSON), &c.Requirements)
	json.Unmarshal([]byte(techStackJSON), &c.TechStack)
	c.Rubric = rubricBytes
	c.Tags = tags
	if c.Tags == nil {
		c.Tags = []string{}
	}
	c.CreatedAt = createdAt.Format(time.RFC3339)

	return &c, nil
}

// GetStarterPackChallenges retrieves challenges for a user's starter pack
func (db *Database) GetStarterPackChallenges(clerkUserID string) ([]ChallengeResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First get the user's starter pack challenge IDs
	var challengeIDsJSON []byte
	err := db.Pool.QueryRow(ctx, `
		SELECT sp.challenge_ids 
		FROM starter_packs sp
		JOIN users u ON u.id = sp.user_id
		WHERE u.clerk_user_id = $1
	`, clerkUserID).Scan(&challengeIDsJSON)

	if err != nil {
		// No starter pack found, return empty
		return []ChallengeResponse{}, nil
	}

	var challengeIDs []string
	if err := json.Unmarshal(challengeIDsJSON, &challengeIDs); err != nil || len(challengeIDs) == 0 {
		return []ChallengeResponse{}, nil
	}

	// Build query to get these challenges
	placeholders := make([]string, len(challengeIDs))
	args := make([]interface{}, len(challengeIDs))
	for i, id := range challengeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT 
			c.id,
			c.title,
			c.description,
			COALESCE(c.description_md, ''),
			c.difficulty,
			c.type,
			c.max_score,
			COALESCE(c.repo_template_url, ''),
			c.requirements,
			c.tech_stack,
			c.estimated_hours,
			c.is_published,
			c.created_at,
			COALESCE(
				(SELECT array_agg(t.name) FROM tags t 
				 JOIN challenge_tags ct ON ct.tag_id = t.id 
				 WHERE ct.challenge_id = c.id),
				ARRAY[]::text[]
			) as tags
		FROM challenges c
		WHERE c.id IN (%s) AND c.is_published = true
		ORDER BY c.difficulty, c.created_at
	`, strings.Join(placeholders, ", "))

	rows, err := db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query starter pack challenges: %w", err)
	}
	defer rows.Close()

	challenges := []ChallengeResponse{}
	for rows.Next() {
		var c ChallengeResponse
		var requirementsJSON, techStackJSON string
		var tags []string
		var createdAt time.Time

		err := rows.Scan(
			&c.ID,
			&c.Title,
			&c.Description,
			&c.DescriptionMD,
			&c.Difficulty,
			&c.Type,
			&c.MaxScore,
			&c.RepoTemplateURL,
			&requirementsJSON,
			&techStackJSON,
			&c.EstimatedHours,
			&c.IsPublished,
			&createdAt,
			&tags,
		)
		if err != nil {
			continue
		}

		json.Unmarshal([]byte(requirementsJSON), &c.Requirements)
		json.Unmarshal([]byte(techStackJSON), &c.TechStack)
		c.Tags = tags
		if c.Tags == nil {
			c.Tags = []string{}
		}
		c.CreatedAt = createdAt.Format(time.RFC3339)

		challenges = append(challenges, c)
	}

	return challenges, nil
}
