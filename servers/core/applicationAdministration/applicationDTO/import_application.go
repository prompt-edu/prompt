package applicationDTO

import (
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
)

// ImportApplicationRequest is the payload for a CSV student import into an application phase.
// The CSV is parsed on the client and sent as structured rows. Columns the lecturer chose to
// import as questions are declared once in NewQuestions and referenced per row by ColumnKey.
type ImportApplicationRequest struct {
	PassStatus   db.PassStatus       `json:"passStatus"`
	NewQuestions []NewImportQuestion `json:"newQuestions"`
	Rows         []ImportRow         `json:"rows"`
}

// NewImportQuestion declares a text application question created from a leftover CSV column.
type NewImportQuestion struct {
	ColumnKey     string `json:"columnKey"`
	Title         string `json:"title"`
	AllowedLength int    `json:"allowedLength"`
}

// ImportRow is a single student together with the answers for the imported question columns.
type ImportRow struct {
	Student studentDTO.CreateStudent `json:"student"`
	Answers []ImportAnswer           `json:"answers"`
}

// ImportAnswer maps a value to an imported question column via its ColumnKey.
type ImportAnswer struct {
	ColumnKey string `json:"columnKey"`
	Answer    string `json:"answer"`
}
