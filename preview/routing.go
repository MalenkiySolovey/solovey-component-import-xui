package preview

import (
	"encoding/json"
	"fmt"

	"github.com/MalenkiySolovey/solovey-backend/storage/model"
	"github.com/MalenkiySolovey/solovey-component-import-xui/mapping"
)

// RoutingProposal carries candidate references for validation by the local
// Routing and DNS owners. It does not admit or persist the candidate.
type RoutingProposal[Unsupported any] struct {
	Item      Item[Unsupported]
	Endpoints []model.Endpoint
	Outbounds []model.Outbound
}

func ProposeRouting[Unsupported any](xrayConfig string) (RoutingProposal[Unsupported], error) {
	endpoints, outbounds, targets, outboundWarnings := mapping.MapXrayOutbounds(xrayConfig)
	mapped, warnings, mappedCount, manualCount := mapping.MapXrayRouting(xrayConfig, targets)
	raw, err := json.MarshalIndent(mapped, "", "  ")
	if err != nil {
		return RoutingProposal[Unsupported]{}, err
	}
	action := ActionCreate
	if xrayConfig == "" || (mappedCount == 0 && manualCount == 0 && len(endpoints) == 0 && len(outbounds) == 0 && !mappedHasDNS(mapped)) {
		action = ActionSkip
	}
	return RoutingProposal[Unsupported]{
		Item: Item[Unsupported]{
			Kind: KindRouting, SrcID: "xrayConfig", SrcTag: "xrayConfig.routing", DstTag: "config",
			Action: action, PreviewJSON: raw, Warnings: append(outboundWarnings, warnings...),
		},
		Endpoints: endpoints,
		Outbounds: outbounds,
	}, nil
}

// RoutingDisabledNotice records intentionally excluded proxy or WARP source rows.
func RoutingDisabledNotice[Unsupported any](xrayConfig string) (Item[Unsupported], bool) {
	endpoints, outbounds, _, _ := mapping.MapXrayOutbounds(xrayConfig)
	if len(endpoints) == 0 && len(outbounds) == 0 {
		return Item[Unsupported]{}, false
	}
	return Item[Unsupported]{
		Kind: KindRouting, SrcID: "xrayConfig", SrcTag: "xrayConfig.outbounds", DstTag: "config",
		Action: ActionSkip, PreviewJSON: json.RawMessage(`null`),
		Warnings: []string{fmt.Sprintf("%d proxy outbound(s) and %d WARP endpoint(s) in the source are not migrated because routing import is disabled; enable \"Include routing\" to migrate them", len(outbounds), len(endpoints))},
	}, true
}

func mappedHasDNS(mapped map[string]any) bool {
	dns, ok := mapped["dns"].(map[string]any)
	return ok && (len(mapping.ToAnySlice(dns["servers"])) > 0 || len(mapping.ToAnySlice(dns["rules"])) > 0)
}
