// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package location

import (
	"context"

	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

// map[string]string handling

// TagsV2 returns location service tags.
func TagsV2(tags tftags.KeyValueTags) map[string]string {
	return tags.Map()
}

// keyValueTagsV2 creates tftags.KeyValueTags from location service tags.
func keyValueTagsV2(ctx context.Context, tags map[string]string) tftags.KeyValueTags {
	return tftags.New(ctx, tags)
}

// getTagsInV2 returns location service tags from Context.
// nil is returned if there are no input tags.
func getTagsInV2(ctx context.Context) map[string]string {
	if inContext, ok := tftags.FromContext(ctx); ok {
		if tags := TagsV2(inContext.TagsIn.UnwrapOrDefault()); len(tags) > 0 {
			return tags
		}
	}

	return nil
}

// setTagsOutV2 sets location service tags in Context.
func setTagsOutV2(ctx context.Context, tags map[string]string) {
	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(keyValueTagsV2(ctx, tags))
	}
}
