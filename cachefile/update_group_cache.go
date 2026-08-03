package cachefile

import (
	"context"
	"encoding/json"

	"github.com/Mantelabs/translaas-sdk-go/models"
)

func (c *CachingClient) updateGroupCache(ctx context.Context, project, group, lang string) error {
	groupData, err := c.inner.GetGroup(ctx, project, group, lang)
	if err != nil {
		return err
	}
	if groupData == nil {
		return nil
	}

	entriesJSON, err := json.Marshal(groupData.Entries)
	if err != nil {
		return err
	}

	var projectToSave *models.TranslationProject
	existing, err := c.cache.GetProject(ctx, project, lang)
	if err != nil {
		return err
	}
	if existing != nil {
		projectToSave = existing
	} else {
		projectToSave = &models.TranslationProject{Groups: make(map[string]json.RawMessage)}
	}
	if projectToSave.Groups == nil {
		projectToSave.Groups = make(map[string]json.RawMessage)
	}
	projectToSave.Groups[group] = entriesJSON

	return c.cache.SaveProject(ctx, project, lang, projectToSave)
}

func (c *CachingClient) tryUpdateGroupCache(ctx context.Context, project, group, lang string) {
	_ = c.updateGroupCache(ctx, project, group, lang)
}
