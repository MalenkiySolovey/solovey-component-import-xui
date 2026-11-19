package preview

import (
	"context"
	"encoding/json"

	"github.com/MalenkiySolovey/solovey-component-import-xui/source"
)

// AddHistorical proposes only aggregated traffic, without a destination write.
func AddHistorical[Unsupported any](ctx context.Context, src *source.Database, plan *Plan[Unsupported]) error {
	if ctx != nil && ctx.Err() != nil {
		return ctx.Err()
	}
	clients, err := src.Clients()
	if err != nil {
		return err
	}
	outbounds, err := src.OutboundTraffics()
	if err != nil {
		return err
	}
	count := 0
	for _, row := range clients {
		if row.Email != "" && (row.Up > 0 || row.Down > 0) {
			count++
		}
	}
	for _, row := range outbounds {
		if row.Tag != "" && (row.Up > 0 || row.Down > 0) {
			count++
		}
	}
	raw, err := json.MarshalIndent(map[string]any{
		"client_traffics":   len(clients),
		"outbound_traffics": len(outbounds),
		"mode":              "aggregated_only",
	}, "", "  ")
	if err != nil {
		return err
	}
	plan.Items = append(plan.Items, Item[Unsupported]{
		Kind:        KindHistory,
		SrcID:       "traffic",
		SrcTag:      "client_traffics/outbound_traffics",
		DstTag:      "stats",
		Action:      ActionCreate,
		PreviewJSON: raw,
		Warnings:    []string{"historical_aggregated_only"},
	})
	plan.Defaults.IncludeHistory = count > 0
	return nil
}
