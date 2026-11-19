// Package preview describes externally sourced import proposals, not authorized changes.
package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/MalenkiySolovey/solovey-component-import-xui/mapping"
)

const (
	KindTLS      = "tls"
	KindInbound  = "inbound"
	KindEndpoint = "endpoint"
	KindClient   = "client"
	KindSetting  = "setting"
	KindAdmin    = "admin"
	KindHistory  = "historical"
	KindRouting  = "routing"

	ActionCreate  = "create"
	ActionMerge   = "merge"
	ActionReplace = "replace"
	ActionSkip    = "skip"
)

// Plan represents a proposal. Unsupported is the destination owner's detail
// type for unavailable capabilities. It does not grant permission to apply.
type Plan[Unsupported any] struct {
	Items    []Item[Unsupported] `json:"items"`
	Defaults PlanDefaults        `json:"defaults"`
	Source   PlanSource          `json:"source"`
}

// Revision excludes the staged path, which is not semantic input. The staged
// file identity and approval for this revision are checked by the Panel.
func (plan Plan[Unsupported]) Revision() (string, error) {
	plan.Source.Path = ""
	raw, err := json.Marshal(plan)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}

type PlanDefaults struct {
	Strategy        string `json:"strategy"`
	IncludeSettings bool   `json:"includeSettings"`
	AdminMode       string `json:"adminMode"`
	OnlyNew         bool   `json:"onlyNew"`
	IncludeHistory  bool   `json:"includeHistory"`
	IncludeRouting  bool   `json:"includeRouting"`
}

type PlanSource struct {
	Path string `json:"path,omitempty"`
	Hash string `json:"hash"`
}

type Item[Unsupported any] struct {
	Classification string                        `json:"classification"`
	Fields         []mapping.FieldClassification `json:"fields,omitempty"`
	Unsupported    []Unsupported                 `json:"unsupported,omitempty"`
	Kind           string                        `json:"kind"`
	SrcID          any                           `json:"srcId"`
	SrcTag         string                        `json:"srcTag,omitempty"`
	DstTag         string                        `json:"dstTag"`
	Action         string                        `json:"action"`
	Conflict       bool                          `json:"conflict"`
	AdminMode      string                        `json:"adminMode,omitempty"`
	PreviewJSON    json.RawMessage               `json:"previewJson"`
	Warnings       []string                      `json:"warnings,omitempty"`
}
