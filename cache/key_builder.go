package cache

import (
	"sort"
	"strconv"
	"strings"
)

const keySeparator = ":"

var defaultKeyBuilder KeyBuilder

// KeyBuilder produces colon-separated keys matching .NET CacheKeyBuilder.
type KeyBuilder struct{}

// EntryKey builds a cache key for a single translation entry.
func EntryKey(group, entry, lang string, n *float64, params map[string]string, project, channel, version string) string {
	return defaultKeyBuilder.EntryKey(group, entry, lang, n, params, project, channel, version)
}

// GroupKey builds a cache key for a translation group payload.
func GroupKey(project, group, lang, format, channel, version string, includeContext *bool) string {
	return defaultKeyBuilder.GroupKey(project, group, lang, format, channel, version, includeContext)
}

// ProjectKey builds a cache key for a full project payload.
func ProjectKey(project, lang, format, channel, version string, includeContext *bool) string {
	return defaultKeyBuilder.ProjectKey(project, lang, format, channel, version, includeContext)
}

// LocalesKey builds a cache key for project locales.
func LocalesKey(project, channel, version string) string {
	return defaultKeyBuilder.LocalesKey(project, channel, version)
}

// OfflineKey builds a cache key for offline ZIP download metadata.
func OfflineKey(project, channel, version string, includeContext *bool) string {
	return defaultKeyBuilder.OfflineKey(project, channel, version, includeContext)
}

// EntryKey builds a cache key for a single translation entry.
func (KeyBuilder) EntryKey(group, entry, lang string, n *float64, params map[string]string, project, channel, version string) string {
	parts := []string{"entry", group, entry, lang}
	if n != nil {
		parts = append(parts, formatCacheNumber(*n))
	}
	parts = append(parts, sortedParamPairs(params)...)
	return appendSnapshotSuffix(parts, project, channel, version, nil)
}

// GroupKey builds a cache key for a translation group payload.
func (KeyBuilder) GroupKey(project, group, lang, format, channel, version string, includeContext *bool) string {
	parts := []string{"group", project, group, lang}
	if strings.TrimSpace(format) != "" {
		parts = append(parts, format)
	}
	return appendSnapshotSuffix(parts, "", channel, version, includeContext)
}

// ProjectKey builds a cache key for a full project payload.
func (KeyBuilder) ProjectKey(project, lang, format, channel, version string, includeContext *bool) string {
	parts := []string{"project", project, lang}
	if strings.TrimSpace(format) != "" {
		parts = append(parts, format)
	}
	return appendSnapshotSuffix(parts, "", channel, version, includeContext)
}

// LocalesKey builds a cache key for project locales.
func (KeyBuilder) LocalesKey(project, channel, version string) string {
	parts := []string{"locales", project}
	return appendSnapshotSuffix(parts, "", channel, version, nil)
}

// OfflineKey builds a cache key for offline ZIP download metadata.
func (KeyBuilder) OfflineKey(project, channel, version string, includeContext *bool) string {
	parts := []string{"offline", project}
	return appendSnapshotSuffix(parts, "", channel, version, includeContext)
}

func appendSnapshotSuffix(parts []string, project, channel, version string, includeContext *bool) string {
	key := strings.Join(parts, keySeparator)
	suffixParts := make([]string, 0, 4)
	if strings.TrimSpace(project) != "" {
		suffixParts = append(suffixParts, "proj="+project)
	}
	if strings.TrimSpace(channel) != "" {
		suffixParts = append(suffixParts, "ch="+channel)
	}
	if strings.TrimSpace(version) != "" {
		suffixParts = append(suffixParts, "v="+version)
	}
	if includeContext != nil {
		if *includeContext {
			suffixParts = append(suffixParts, "ic=1")
		} else {
			suffixParts = append(suffixParts, "ic=0")
		}
	}
	if len(suffixParts) == 0 {
		return key
	}
	return key + keySeparator + strings.Join(suffixParts, keySeparator)
}

func sortedParamPairs(params map[string]string) []string {
	if len(params) == 0 {
		return nil
	}
	keys := make([]string, 0, len(params))
	for key := range params {
		if strings.TrimSpace(key) == "" {
			continue
		}
		val := params[key]
		if strings.TrimSpace(val) == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.ToLower(keys[i]) < strings.ToLower(keys[j])
	})
	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, strings.ToLower(key)+"="+params[key])
	}
	return pairs
}

// formatCacheNumber matches Python format(n, ".15g") / .NET invariant general format.
func formatCacheNumber(n float64) string {
	return strconv.FormatFloat(n, 'g', 15, 64)
}
