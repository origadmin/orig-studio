package dal

import (
	"context"
	"log/slog"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/category"
	"origadmin/application/origstudio/internal/dal/entity/tag"
)

// catSpec describes a category to seed: a module root (ParentID nil) or a
// video-genre child (ParentID = video root id).
type catSpec struct {
	Name        string
	Slug        string
	Icon        string
	Color       string
	Sequence    int
	Description string
	NameI18n    map[string]string
	IsGlobal    bool
	ParentID    *int64 // nil => module root (parent_id IS NULL)
}

// SeedCategories builds the shared taxonomy from taxonomy.yaml (D3):
//   - three module roots (video / music / article) as anchors. The category
//     table is shared across modules (edges to Media/Article/Channel), so each
//     module gets a root and a category is anchored to its module by tree
//     position — not by any media.type field.
//   - the video root carries the genre/form subtree (教程/游戏/连续剧/电影/…/其他),
//     including the §4.2 additions (drama/movie/variety/anime/mv).
//
// Idempotent by slug: re-running never duplicates rows, and a stale live DB is
// reconciled to the taxonomy (legacy slugs education→tutorial, technology→tech
// are renamed BEFORE the tree upsert so the new slug never collides).
func SeedCategories(ctx context.Context, client *entity.Client) error {
	spec, err := seedTaxonomy()
	if err != nil {
		return err
	}

	// Legacy slug migration FIRST: on a live DB the video root already exists,
	// so we can rename education→tutorial / technology→tech before the tree
	// upsert creates those slugs (avoiding a unique-slug collision). No-op on a
	// fresh DB (video root absent → IsNotFound → skip).
	if videoID, err := client.Category.Query().Where(category.SlugEQ("video")).OnlyID(ctx); err == nil {
		if err := renameLegacyCategory(ctx, client, "education", "tutorial", "教程", videoID); err != nil {
			return err
		}
		if err := renameLegacyCategory(ctx, client, "technology", "tech", "科技", videoID); err != nil {
			return err
		}
	} else if !entity.IsNotFound(err) {
		return err
	}

	var seeded int
	if err := upsertCategoryTree(ctx, client, spec.Categories, nil, &seeded); err != nil {
		return err
	}

	slog.Info("Successfully seeded category taxonomy from taxonomy.yaml", "categories", seeded)
	return nil
}

// upsertCategoryTree recursively upserts a taxonomy subtree. Roots pass nil
// ParentID (module anchors), children pass their parent's DB id (tree position
// is the module anchor, see BUG-145).
func upsertCategoryTree(ctx context.Context, client *entity.Client, nodes []*catSpecNode, parentID *int64, counter *int) error {
	for _, n := range nodes {
		isGlobal := true
		if n.IsGlobal != nil {
			isGlobal = *n.IsGlobal
		}
		id, err := upsertCategory(ctx, client, catSpec{
			Name: n.Name, Slug: n.Slug, Icon: n.Icon, Color: n.Color,
			Sequence: n.Sequence, Description: n.Description,
			IsGlobal: isGlobal, ParentID: parentID,
		})
		if err != nil {
			return err
		}
		*counter++
		childID := id
		if err := upsertCategoryTree(ctx, client, n.Children, &childID, counter); err != nil {
			return err
		}
	}
	return nil
}

// upsertCategory creates or updates a category by slug. Safe to call on every
// startup: existing rows are reconciled to the target state, missing rows are
// created. Never duplicates.
func upsertCategory(ctx context.Context, client *entity.Client, spec catSpec) (int64, error) {
	existing, err := client.Category.Query().Where(category.SlugEQ(spec.Slug)).Only(ctx)
	if entity.IsNotFound(err) {
		create := client.Category.Create().
			SetName(spec.Name).
			SetSlug(spec.Slug).
			SetIcon(spec.Icon).
			SetColor(spec.Color).
			SetSequence(spec.Sequence).
			SetIsGlobal(spec.IsGlobal)
		if spec.Description != "" {
			create.SetDescription(spec.Description)
		}
		if spec.NameI18n != nil {
			create.SetNameI18n(spec.NameI18n)
		}
		if spec.ParentID != nil {
			create.SetParentID(*spec.ParentID)
		}
		c, err := create.Save(ctx)
		if err != nil {
			return 0, err
		}
		return c.ID, nil
	}
	if err != nil {
		return 0, err
	}

	upd := existing.Update().
		SetName(spec.Name).
		SetIcon(spec.Icon).
		SetColor(spec.Color).
		SetSequence(spec.Sequence).
		SetIsGlobal(spec.IsGlobal)
	if spec.Description != "" {
		upd.SetDescription(spec.Description)
	}
	if spec.NameI18n != nil {
		upd.SetNameI18n(spec.NameI18n)
	}
	if spec.ParentID != nil {
		upd.SetParentID(*spec.ParentID)
	} else {
		// Module root: parent_id must be NULL. It carries an FK to
		// content_categories(id), so writing 0 would violate the constraint.
		upd.ClearParent()
	}
	if _, err := upd.Save(ctx); err != nil {
		return 0, err
	}
	return existing.ID, nil
}

// renameLegacyCategory migrates an old slug to the agreed taxonomy on an
// existing DB. It is a no-op when the old slug is absent (fresh DB), so the
// seed stays idempotent regardless of run order.
func renameLegacyCategory(ctx context.Context, client *entity.Client, oldSlug, newSlug, newName string, parentID int64) error {
	old, err := client.Category.Query().Where(category.SlugEQ(oldSlug)).Only(ctx)
	if entity.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	// If the target slug (or a row already carrying the target name) already
	// exists, the legacy row is redundant — delete it instead of renaming.
	// Renaming would collide with content_categories_name_key / slug unique
	// constraints on a DB where both old and new slugs coexist (e.g. a
	// partially-seeded taxonomy), crashing startup.
	exists, err := client.Category.Query().
		Where(category.Or(category.SlugEQ(newSlug), category.NameEQ(newName))).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return client.Category.DeleteOne(old).Exec(ctx)
	}
	_, err = old.Update().
		SetSlug(newSlug).
		SetName(newName).
		SetParentID(parentID).
		Save(ctx)
	return err
}

// SeedTags seeds tags from taxonomy.yaml by slug upsert (D4). The old
// `count>0 → skip` first-run logic is gone: taxonomy.yaml is the source of
// truth and tags are corrected on every startup (rows not listed are kept).
func SeedTags(ctx context.Context, client *entity.Client) error {
	spec, err := seedTaxonomy()
	if err != nil {
		return err
	}

	for _, t := range spec.Tags {
		if err := upsertTag(ctx, client, *t); err != nil {
			slog.Error("failed to seed tag", "title", t.Title, "err", err)
			return err
		}
	}

	slog.Info("Successfully seeded tags from taxonomy.yaml", "tags", len(spec.Tags))
	return nil
}

// upsertTag creates or updates a tag by slug. Idempotent: missing rows are
// created, existing rows are reconciled to the taxonomy (title/color/
// description). Status is left untouched (rows not in the config are kept).
func upsertTag(ctx context.Context, client *entity.Client, spec tagSpec) error {
	existing, err := client.Tag.Query().Where(tag.SlugEQ(spec.Slug)).Only(ctx)
	if entity.IsNotFound(err) {
		create := client.Tag.Create().
			SetTitle(spec.Title).
			SetSlug(spec.Slug)
		if spec.Color != "" {
			create.SetColor(spec.Color)
		}
		if spec.Description != "" {
			create.SetDescription(spec.Description)
		}
		_, err = create.Save(ctx)
		return err
	}
	if err != nil {
		return err
	}

	upd := existing.Update().SetTitle(spec.Title)
	if spec.Color != "" {
		upd.SetColor(spec.Color)
	}
	if spec.Description != "" {
		upd.SetDescription(spec.Description)
	}
	_, err = upd.Save(ctx)
	return err
}
