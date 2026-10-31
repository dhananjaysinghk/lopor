package datasynth

import (
	"time"
)

// PrivacyLevel dictates differential privacy epsilon bounds and strictness of PII masking.
type PrivacyLevel string

const (
	PrivacyPublic     PrivacyLevel = "PUBLIC"     // Epsilon ~ 2.0 (Low noise, high statistical fidelity)
	PrivacyInternal   PrivacyLevel = "INTERNAL"   // Epsilon ~ 1.0 (Balanced utility and privacy)
	PrivacyConfidential PrivacyLevel = "CONFIDENTIAL" // Epsilon ~ 0.5 (Strict Laplace noise injection)
	PrivacyRestricted PrivacyLevel = "RESTRICTED" // Epsilon ~ 0.1 (Maximum zero-knowledge differential noise)
)

// FieldDataType specifies primitive or synthetic semantic type for structured columns.
type FieldDataType string

const (
	FieldTypeUUID         FieldDataType = "UUID"
	FieldTypeEmail        FieldDataType = "EMAIL"
	FieldTypeName         FieldDataType = "FULL_NAME"
	FieldTypePhone        FieldDataType = "PHONE"
	FieldTypeIntegerRange FieldDataType = "INTEGER_RANGE"
	FieldTypeFloatNormal  FieldDataType = "FLOAT_NORMAL"
	FieldTypeCategorical  FieldDataType = "CATEGORICAL"
	FieldTypeTimestamp    FieldDataType = "TIMESTAMP"
	FieldTypeFreeText     FieldDataType = "FREE_TEXT"
)

// SyntheticFieldSchema describes an individual property or table column to synthesize.
type SyntheticFieldSchema struct {
	Name            string        `json:"name"`
	Type            FieldDataType `json:"type"`
	MinValue        float64       `json:"min_value,omitempty"`
	MaxValue        float64       `json:"max_value,omitempty"`
	Mean            float64       `json:"mean,omitempty"`
	StdDev          float64       `json:"std_dev,omitempty"`
	Categories      []string      `json:"categories,omitempty"`
	Weights         []float64     `json:"weights,omitempty"` // Probability distribution for categories
	IsNullable      bool          `json:"is_nullable,omitempty"`
	NullProbability float64       `json:"null_probability,omitempty"`
}

// SyntheticDatasetConfig configures the volume, privacy tier, and generation seed.
type SyntheticDatasetConfig struct {
	RowCount          int                    `json:"row_count"`
	PrivacyLevel      PrivacyLevel           `json:"privacy_level"`
	Epsilon           float64                `json:"epsilon,omitempty"` // Differential privacy budget (default 1.0)
	EnablePIIMasking  bool                   `json:"enable_pii_masking"`
	DomainContext     string                 `json:"domain_context,omitempty"` // e.g. "Fintech customer transaction dataset"
	Fields            []SyntheticFieldSchema `json:"fields"`
}

// GenerateDatasetRequest defines request parameters to spawn synthetic generation.
type GenerateDatasetRequest struct {
	WorkspaceID string                 `json:"workspace_id"`
	Title       string                 `json:"title"`
	Config      SyntheticDatasetConfig `json:"config"`
}

// PrivacyMetricReport measures the mathematical differential privacy and leakage risk.
type PrivacyMetricReport struct {
	EpsilonUsed              float64 `json:"epsilon_used"`
	ReidentificationRiskScore float64 `json:"reidentification_risk_score"` // 0.0 - 100.0% (Lower is better)
	PIIResidualLeaksCount    int     `json:"pii_residual_leaks_count"`
	StatisticalCorrelation   float64 `json:"statistical_correlation"`     // 0.0 - 1.0 (Higher preserves real distributions)
	IsGDPRCompliant          bool    `json:"is_gdpr_compliant"`
}

// GeneratedDataset encapsulates the generated records, schema metadata, and audit metrics.
type GeneratedDataset struct {
	DatasetID       string                   `json:"dataset_id"`
	WorkspaceID     string                   `json:"workspace_id"`
	Title           string                   `json:"title"`
	RowCount        int                      `json:"row_count"`
	PrivacyLevel    PrivacyLevel             `json:"privacy_level"`
	Records         []map[string]interface{} `json:"records"`
	Metrics         PrivacyMetricReport      `json:"metrics"`
	GeneratedAt     time.Time                `json:"generated_at"`
	ExecutionTimeMs int64                    `json:"execution_time_ms"`
}

