package agents

import (
	"context"
	"log"

	"github.com/sin-engine/ai-assistant/memory"
	"github.com/sin-engine/ai-assistant/tools"
)

// Action represents a planned action
type Action struct {
	Type   string                 `json:"type"` // search, scan, exploit, breach_check, explain, learn
	Tool   string                 `json:"tool"`
	Params map[string]interface{} `json:"params"`
	Reason string                 `json:"reason"`
	Order  int                    `json:"order"`
}

// Plan represents a plan of actions
type Plan struct {
	Goal    string   `json:"goal"`
	Actions []Action `json:"actions"`
	Reason  string   `json:"reason"`
}

// Planner creates action plans based on analysis
type Planner struct {
	availableTools []tools.Tool
}

// NewPlanner creates a new planner
func NewPlanner(availableTools []tools.Tool) *Planner {
	return &Planner{
		availableTools: availableTools,
	}
}

// CreatePlan creates a plan based on analysis and knowledge
func (p *Planner) CreatePlan(ctx context.Context, analysis *Analysis, knowledge []memory.KnowledgeItem, context map[string]interface{}) (*Plan, error) {
	plan := &Plan{
		Goal:    analysis.Intent,
		Actions: []Action{},
		Reason:  "",
	}

	// Route based on intent
	switch analysis.Intent {
	case "greeting":
		plan.Actions = append(plan.Actions, Action{
			Type:   "respond",
			Tool:   "chat",
			Reason: "Respond to greeting",
		})

	case "help":
		plan.Actions = append(plan.Actions, Action{
			Type:   "respond",
			Tool:   "chat",
			Reason: "Provide help information",
		})

	case "explain":
		plan.Actions = append(plan.Actions, Action{
			Type: "respond",
			Tool: "chat",
			Params: map[string]interface{}{
				"topic": analysis.Topics,
			},
			Reason: "Explain the topic",
		})

	case "search":
		query := getStringFromContext(context, "query", "")
		plan.Actions = append(plan.Actions, Action{
			Type:   "execute",
			Tool:   "search",
			Params: map[string]interface{}{"query": query},
			Reason: "Search for resources",
		})

	case "dork":
		dork := getStringFromContext(context, "dork", "")
		plan.Actions = append(plan.Actions, Action{
			Type:   "execute",
			Tool:   "dork",
			Params: map[string]interface{}{"dork": dork},
			Reason: "Search using dorks",
		})

	case "scan":
		target := getStringFromContext(context, "target", "")
		scanType := getStringFromContext(context, "scan_type", "basic")
		plan.Actions = append(plan.Actions, Action{
			Type:   "execute",
			Tool:   "scanner",
			Params: map[string]interface{}{"target": target, "scan_type": scanType},
			Reason: "Scan target for vulnerabilities",
		})

	case "breach_check":
		email := getStringFromContext(context, "email", "")
		plan.Actions = append(plan.Actions, Action{
			Type:   "execute",
			Tool:   "breach",
			Params: map[string]interface{}{"email": email},
			Reason: "Check for breaches",
		})

	case "exploit":
		cve := getStringFromContext(context, "cve", "")
		plan.Actions = append(plan.Actions, Action{
			Type:   "execute",
			Tool:   "exploiter",
			Params: map[string]interface{}{"cve": cve},
			Reason: "Search for exploits",
		})

	case "learn":
		topic := getStringFromContext(context, "topic", "")
		plan.Actions = append(plan.Actions, Action{
			Type:   "search",
			Tool:   "search",
			Params: map[string]interface{}{"query": topic, "category": "learning"},
			Reason: "Find learning resources",
		})

		if len(knowledge) > 0 {
			plan.Actions = append(plan.Actions, Action{
				Type:   "respond",
				Tool:   "chat",
				Reason: "Use existing knowledge",
			})
		}

	case "chat":
		plan.Actions = append(plan.Actions, Action{
			Type:   "respond",
			Tool:   "chat",
			Reason: "Continue conversation",
		})

	default:
		plan.Actions = append(plan.Actions, Action{
			Type:   "respond",
			Tool:   "chat",
			Reason: "Default response",
		})
	}

	log.Printf("Created plan with %d actions for intent: %s", len(plan.Actions), analysis.Intent)

	return plan, nil
}

func getStringFromContext(context map[string]interface{}, key, defaultValue string) string {
	if context == nil {
		return defaultValue
	}
	if val, ok := context[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}
