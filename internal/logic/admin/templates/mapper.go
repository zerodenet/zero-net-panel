package templates

import (
	"github.com/zero-net-panel/zero-net-panel/internal/repository"
	"github.com/zero-net-panel/zero-net-panel/internal/types"
)

func toTemplateSummary(t repository.SubscriptionTemplate) types.SubscriptionTemplateSummary {
	summary := types.SubscriptionTemplateSummary{
		ID:              t.ID,
		Name:            t.Name,
		Description:     t.Description,
		ClientType:      t.ClientType,
		Format:          t.Format,
		Content:         t.Content,
		Variables:       cloneVariables(t.Variables),
		IsDefault:       t.IsDefault,
		Version:         t.Version,
		UpdatedAt:       t.UpdatedAt.Unix(),
		LastPublishedBy: t.LastPublishedBy,
	}
	if t.PublishedAt != nil {
		summary.PublishedAt = t.PublishedAt.Unix()
	}
	return summary
}

func toHistoryEntry(h repository.SubscriptionTemplateHistory) types.SubscriptionTemplateHistoryEntry {
	entry := types.SubscriptionTemplateHistoryEntry{
		Version:     h.Version,
		Changelog:   h.Changelog,
		PublishedAt: h.PublishedAt.Unix(),
		PublishedBy: h.PublishedBy,
		Variables:   cloneVariables(h.Variables),
	}
	return entry
}

func cloneVariables(vars map[string]repository.TemplateVariable) map[string]types.TemplateVariable {
	if vars == nil {
		return nil
	}
        cloned := make(map[string]types.TemplateVariable, len(vars))
        for k, v := range vars {
                cloned[k] = types.TemplateVariable{
                        ValueType:   v.ValueType,
                        Required:    v.Required,
                        Description: v.Description,
                        DefaultValue: v.DefaultValue,
                }
        }
	return cloned
}

func toRepositoryVariables(vars map[string]types.TemplateVariable) map[string]repository.TemplateVariable {
	if vars == nil {
		return nil
	}
        cloned := make(map[string]repository.TemplateVariable, len(vars))
        for k, v := range vars {
                cloned[k] = repository.TemplateVariable{
                        ValueType:   v.ValueType,
                        Required:    v.Required,
                        Description: v.Description,
                        DefaultValue: v.DefaultValue,
                }
        }
	return cloned
}
