package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	entsql "entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/imagecreationasset"
	"github.com/Wei-Shaw/sub2api/ent/imagecreationtemplate"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type imageCreationRepository struct {
	client *dbent.Client
}

func NewImageCreationRepository(client *dbent.Client) service.ImageCreationRepository {
	return &imageCreationRepository{client: client}
}

func (r *imageCreationRepository) StoreAsset(ctx context.Context, asset *service.ImageCreationAsset) (*service.ImageCreationAsset, bool, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Client().ExecContext(ctx, `
		INSERT INTO image_creation_assets
			(id, content, content_type, byte_size, width, height, source_type, source_provider, source_model, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), $10)
		ON CONFLICT (id) DO NOTHING`,
		asset.ID, asset.Content, asset.ContentType, asset.ByteSize, asset.Width, asset.Height,
		asset.SourceType, asset.SourceProvider, asset.SourceModel, asset.CreatedBy,
	)
	if err != nil {
		return nil, false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	created := rows == 1
	if created {
		if err := createImageCreationChangeLog(ctx, tx.Client(), asset.CreatedBy, "asset_create", "asset", asset.ID, map[string]any{
			"content_type": asset.ContentType, "byte_size": asset.ByteSize,
		}); err != nil {
			return nil, false, err
		}
	}
	entity, err := tx.Client().ImageCreationAsset.Get(ctx, asset.ID)
	if err != nil {
		return nil, false, translatePersistenceError(err, service.ErrImageCreationNotFound, nil)
	}
	if err := tx.Commit(); err != nil {
		return nil, false, err
	}
	return imageCreationAssetEntityToService(entity, true), created, nil
}

func (r *imageCreationRepository) GetAsset(ctx context.Context, id string, withContent bool) (*service.ImageCreationAsset, error) {
	q := r.client.ImageCreationAsset.Query().Where(imagecreationasset.IDEQ(id))
	if !withContent {
		q.Select(
			imagecreationasset.FieldID,
			imagecreationasset.FieldContentType,
			imagecreationasset.FieldByteSize,
			imagecreationasset.FieldWidth,
			imagecreationasset.FieldHeight,
			imagecreationasset.FieldSourceType,
			imagecreationasset.FieldSourceProvider,
			imagecreationasset.FieldSourceModel,
			imagecreationasset.FieldCreatedBy,
			imagecreationasset.FieldCreatedAt,
		)
	}
	entity, err := q.Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrImageCreationNotFound, nil)
	}
	return imageCreationAssetEntityToService(entity, withContent), nil
}

func (r *imageCreationRepository) CreateTemplate(ctx context.Context, template *service.ImageCreationTemplate) (*service.ImageCreationTemplate, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := ensureImageCreationAssetExists(ctx, tx.Client(), template.DraftCoverAssetID); err != nil {
		return nil, err
	}
	builder := tx.Client().ImageCreationTemplate.Create().
		SetState(template.State).
		SetDraftData(template.DraftData).
		SetRevision(template.Revision).
		SetCreatedBy(template.CreatedBy).
		SetUpdatedBy(template.UpdatedBy)
	if template.DraftCoverAssetID != nil {
		builder.SetDraftCoverAssetID(*template.DraftCoverAssetID)
	}
	entity, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := createImageCreationChangeLog(ctx, tx.Client(), template.CreatedBy, "create", "template", strconv.FormatInt(entity.ID, 10), map[string]any{"revision": entity.Revision}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return imageCreationTemplateEntityToService(entity), nil
}

func (r *imageCreationRepository) GetTemplate(ctx context.Context, id int64, publishedOnly bool) (*service.ImageCreationTemplate, error) {
	q := r.client.ImageCreationTemplate.Query().Where(imagecreationtemplate.IDEQ(id))
	if publishedOnly {
		q.Where(imagecreationtemplate.StateEQ(service.ImageCreationTemplateStatePublished))
	}
	entity, err := q.Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrImageCreationNotFound, nil)
	}
	return imageCreationTemplateEntityToService(entity), nil
}

func (r *imageCreationRepository) ListTemplates(ctx context.Context, userID int64, params pagination.PaginationParams, filters service.ImageCreationTemplateListFilters, admin bool) ([]service.ImageCreationTemplate, *pagination.PaginationResult, error) {
	q := r.client.ImageCreationTemplate.Query()
	if admin {
		if filters.State != "" {
			q.Where(imagecreationtemplate.StateEQ(filters.State))
		}
	} else {
		q.Where(imagecreationtemplate.StateEQ(service.ImageCreationTemplateStatePublished))
	}
	if filters.Query != "" {
		field := imagecreationtemplate.FieldPublishedData
		if admin {
			field = imagecreationtemplate.FieldDraftData
		}
		q.Where(func(s *entsql.Selector) {
			s.Where(entsql.Or(
				imageCreationJSONContainsFold(field, "title", filters.Query),
				imageCreationJSONContainsFold(field, "summary", filters.Query),
			))
		})
	}
	if filters.Category != "" {
		q.Where(func(s *entsql.Selector) {
			s.Where(sqljson.ValueEQ(imagecreationtemplate.FieldPublishedData, filters.Category, sqljson.Path("category")))
		})
	}
	if filters.Tag != "" {
		q.Where(func(s *entsql.Selector) {
			s.Where(sqljson.ValueContains(imagecreationtemplate.FieldPublishedData, []string{filters.Tag}, sqljson.Path("tags")))
		})
	}
	if filters.Home {
		q.Where(imagecreationtemplate.HomePositionNotNil())
	}
	if !admin && (filters.Favorite || filters.Recent) {
		q.Where(func(s *entsql.Selector) {
			state := entsql.Table("image_creation_user_template_states").As("icts")
			s.Join(state).On(s.C(imagecreationtemplate.FieldID), state.C("template_id"))
			s.Where(entsql.EQ(state.C("user_id"), userID))
			if filters.Favorite {
				s.Where(entsql.NotNull(state.C("favorited_at")))
			}
			if filters.Recent {
				s.Where(entsql.NotNull(state.C("last_used_at")))
			}
		})
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, nil, err
	}
	itemsQuery := q.Offset(params.Offset()).Limit(params.Limit())
	switch {
	case filters.Home:
		itemsQuery.Order(dbent.Asc(imagecreationtemplate.FieldHomePosition), dbent.Asc(imagecreationtemplate.FieldID))
	case filters.Recent && !admin:
		itemsQuery.Order(func(s *entsql.Selector) {
			s.OrderExpr(entsql.Expr("icts.last_used_at DESC NULLS LAST"))
			s.OrderBy(entsql.Desc(s.C(imagecreationtemplate.FieldID)))
		})
	case admin:
		itemsQuery.Order(dbent.Desc(imagecreationtemplate.FieldUpdatedAt), dbent.Desc(imagecreationtemplate.FieldID))
	default:
		itemsQuery.Order(dbent.Desc(imagecreationtemplate.FieldPublishedAt), dbent.Desc(imagecreationtemplate.FieldID))
	}
	entities, err := itemsQuery.All(ctx)
	if err != nil {
		return nil, nil, err
	}
	items := imageCreationTemplateEntitiesToService(entities)
	if !admin && len(items) > 0 {
		if err := r.attachImageCreationUserStates(ctx, userID, items); err != nil {
			return nil, nil, err
		}
	}
	return items, paginationResultFromTotal(int64(total), params), nil
}

func (r *imageCreationRepository) UpdateDraft(ctx context.Context, id int64, revision int, doc service.ImageCreationTemplateDocument, coverAssetID *string, actorID int64) (*service.ImageCreationTemplate, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := ensureImageCreationAssetExists(ctx, tx.Client(), coverAssetID); err != nil {
		return nil, err
	}
	update := tx.Client().ImageCreationTemplate.Update().Where(
		imagecreationtemplate.IDEQ(id),
		imagecreationtemplate.RevisionEQ(revision),
		imagecreationtemplate.StateNEQ(service.ImageCreationTemplateStateArchived),
	).SetDraftData(doc).SetUpdatedBy(actorID).AddRevision(1)
	if coverAssetID == nil {
		update.ClearDraftCoverAssetID()
	} else {
		update.SetDraftCoverAssetID(*coverAssetID)
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	if updated != 1 {
		return nil, imageCreationTemplateCASFailure(ctx, tx.Client(), id)
	}
	if err := createImageCreationChangeLog(ctx, tx.Client(), actorID, "update", "template", strconv.FormatInt(id, 10), map[string]any{"revision": revision + 1}); err != nil {
		return nil, err
	}
	entity, err := tx.Client().ImageCreationTemplate.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return imageCreationTemplateEntityToService(entity), nil
}

func (r *imageCreationRepository) PublishTemplate(ctx context.Context, id int64, revision int, actorID int64) (*service.ImageCreationTemplate, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	entity, err := tx.Client().ImageCreationTemplate.Query().Where(imagecreationtemplate.IDEQ(id)).ForUpdate().Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrImageCreationNotFound, nil)
	}
	if entity.Revision != revision || entity.State == service.ImageCreationTemplateStateArchived {
		return nil, service.ErrImageCreationConflict
	}
	if entity.DraftCoverAssetID == nil {
		return nil, service.ErrImageCreationCoverRequired
	}
	now := time.Now()
	publishedVersion := entity.PublishedVersion + 1
	updated, err := tx.Client().ImageCreationTemplate.UpdateOneID(id).
		SetState(service.ImageCreationTemplateStatePublished).
		SetPublishedData(&entity.DraftData).
		SetPublishedCoverAssetID(*entity.DraftCoverAssetID).
		SetPublishedVersion(publishedVersion).
		SetPublishedAt(now).
		SetUpdatedBy(actorID).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := createImageCreationChangeLog(ctx, tx.Client(), actorID, "publish", "template", strconv.FormatInt(id, 10), map[string]any{"published_version": publishedVersion}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return imageCreationTemplateEntityToService(updated), nil
}

func (r *imageCreationRepository) ArchiveTemplate(ctx context.Context, id, actorID int64) (*service.ImageCreationTemplate, error) {
	return r.changeImageCreationTemplateState(ctx, id, actorID, service.ImageCreationTemplateStateArchived)
}

func (r *imageCreationRepository) RestoreTemplate(ctx context.Context, id, actorID int64) (*service.ImageCreationTemplate, error) {
	return r.changeImageCreationTemplateState(ctx, id, actorID, service.ImageCreationTemplateStateDraft)
}

func (r *imageCreationRepository) changeImageCreationTemplateState(ctx context.Context, id, actorID int64, next string) (*service.ImageCreationTemplate, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	entity, err := tx.Client().ImageCreationTemplate.Query().Where(imagecreationtemplate.IDEQ(id)).ForUpdate().Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrImageCreationNotFound, nil)
	}
	if entity.State == next {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		return imageCreationTemplateEntityToService(entity), nil
	}
	if next == service.ImageCreationTemplateStateDraft && entity.State != service.ImageCreationTemplateStateArchived {
		return nil, service.ErrImageCreationConflict
	}
	update := tx.Client().ImageCreationTemplate.UpdateOneID(id).SetState(next).SetUpdatedBy(actorID)
	action := "restore"
	if next == service.ImageCreationTemplateStateArchived {
		update.ClearHomePosition()
		action = "archive"
	}
	updated, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}
	if err := createImageCreationChangeLog(ctx, tx.Client(), actorID, action, "template", strconv.FormatInt(id, 10), map[string]any{"state": next}); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return imageCreationTemplateEntityToService(updated), nil
}

func (r *imageCreationRepository) GetHomeFeatured(ctx context.Context) (*service.ImageCreationHomeFeatured, error) {
	return getImageCreationHomeFeatured(ctx, r.client)
}

func (r *imageCreationRepository) ReplaceHomeFeatured(ctx context.Context, etag string, templateIDs []int64, actorID int64) (*service.ImageCreationHomeFeatured, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Client().ExecContext(ctx, "SELECT pg_advisory_xact_lock(480661784923682723)"); err != nil {
		return nil, err
	}
	current, err := getImageCreationHomeFeatured(ctx, tx.Client())
	if err != nil {
		return nil, err
	}
	if current.ETag != etag {
		return nil, service.ErrImageCreationConflict
	}
	if len(templateIDs) > 0 {
		count, err := tx.Client().ImageCreationTemplate.Query().Where(
			imagecreationtemplate.IDIn(templateIDs...),
			imagecreationtemplate.StateEQ(service.ImageCreationTemplateStatePublished),
		).Count(ctx)
		if err != nil {
			return nil, err
		}
		if count != len(templateIDs) {
			return nil, service.ErrImageCreationNotFound
		}
	}
	if _, err := tx.Client().ImageCreationTemplate.Update().Where(imagecreationtemplate.HomePositionNotNil()).ClearHomePosition().Save(ctx); err != nil {
		return nil, err
	}
	for i, id := range templateIDs {
		if err := tx.Client().ImageCreationTemplate.UpdateOneID(id).SetHomePosition(int16(i + 1)).SetUpdatedBy(actorID).Exec(ctx); err != nil {
			return nil, err
		}
	}
	if err := createImageCreationChangeLog(ctx, tx.Client(), actorID, "home_update", "home", "featured", map[string]any{"template_ids": templateIDs}); err != nil {
		return nil, err
	}
	updated, err := getImageCreationHomeFeatured(ctx, tx.Client())
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *imageCreationRepository) SetFavorite(ctx context.Context, userID, templateID int64, favorite bool) error {
	exists, err := r.client.ImageCreationTemplate.Query().Where(
		imagecreationtemplate.IDEQ(templateID),
		imagecreationtemplate.StateEQ(service.ImageCreationTemplateStatePublished),
	).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return service.ErrImageCreationNotFound
	}
	if favorite {
		_, err = r.client.ExecContext(ctx, `
			INSERT INTO image_creation_user_template_states (user_id, template_id, favorited_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (user_id, template_id) DO UPDATE SET favorited_at = EXCLUDED.favorited_at`, userID, templateID)
		return err
	}
	_, err = r.client.ExecContext(ctx, `
		WITH deleted AS (
			DELETE FROM image_creation_user_template_states
			WHERE user_id = $1 AND template_id = $2 AND last_used_at IS NULL
			RETURNING 1
		)
		UPDATE image_creation_user_template_states
		SET favorited_at = NULL
		WHERE user_id = $1 AND template_id = $2 AND last_used_at IS NOT NULL`, userID, templateID)
	return err
}

func (r *imageCreationRepository) ApplyTemplate(ctx context.Context, userID, templateID int64, publishedVersion int) (*service.ImageCreationTemplateApplication, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	entity, err := tx.Client().ImageCreationTemplate.Query().Where(
		imagecreationtemplate.IDEQ(templateID),
		imagecreationtemplate.StateEQ(service.ImageCreationTemplateStatePublished),
	).Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrImageCreationNotFound, nil)
	}
	if entity.PublishedVersion != publishedVersion || entity.PublishedData == nil {
		return nil, service.ErrImageCreationConflict
	}
	if _, err := tx.Client().ExecContext(ctx, `
		INSERT INTO image_creation_user_template_states (user_id, template_id, last_used_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id, template_id) DO UPDATE SET last_used_at = EXCLUDED.last_used_at`, userID, templateID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &service.ImageCreationTemplateApplication{
		TemplateID: templateID, PublishedVersion: entity.PublishedVersion,
		Prompt: entity.PublishedData.Prompt, Defaults: entity.PublishedData.Defaults, InputMode: entity.PublishedData.InputMode,
	}, nil
}

func getImageCreationHomeFeatured(ctx context.Context, client *dbent.Client) (*service.ImageCreationHomeFeatured, error) {
	entities, err := client.ImageCreationTemplate.Query().Where(
		imagecreationtemplate.StateEQ(service.ImageCreationTemplateStatePublished),
		imagecreationtemplate.HomePositionNotNil(),
	).Order(dbent.Asc(imagecreationtemplate.FieldHomePosition)).All(ctx)
	if err != nil {
		return nil, err
	}
	items := imageCreationTemplateEntitiesToService(entities)
	ids := make([]int64, 0, len(items))
	parts := make([]string, 0, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
		parts = append(parts, fmt.Sprintf("%d:%d:%d", items[i].ID, items[i].PublishedVersion, *items[i].HomePosition))
	}
	hash := sha256.Sum256([]byte(strings.Join(parts, ",")))
	return &service.ImageCreationHomeFeatured{ETag: `"` + hex.EncodeToString(hash[:]) + `"`, TemplateIDs: ids, Templates: items}, nil
}

func (r *imageCreationRepository) attachImageCreationUserStates(ctx context.Context, userID int64, items []service.ImageCreationTemplate) error {
	ids := make([]int64, 0, len(items))
	byID := make(map[int64]*service.ImageCreationTemplate, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
		byID[items[i].ID] = &items[i]
	}
	rows, err := r.client.QueryContext(ctx, `
		SELECT template_id, favorited_at, last_used_at
		FROM image_creation_user_template_states
		WHERE user_id = $1 AND template_id = ANY($2)`, userID, pq.Array(ids))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var favorite, recent sql.NullTime
		if err := rows.Scan(&id, &favorite, &recent); err != nil {
			return err
		}
		item := byID[id]
		if item == nil {
			continue
		}
		if favorite.Valid {
			item.FavoritedAt = &favorite.Time
		}
		if recent.Valid {
			item.LastUsedAt = &recent.Time
		}
	}
	return rows.Err()
}

func ensureImageCreationAssetExists(ctx context.Context, client *dbent.Client, id *string) error {
	if id == nil {
		return nil
	}
	exists, err := client.ImageCreationAsset.Query().Where(imagecreationasset.IDEQ(*id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return service.ErrImageCreationNotFound
	}
	return nil
}

func imageCreationTemplateCASFailure(ctx context.Context, client *dbent.Client, id int64) error {
	exists, err := client.ImageCreationTemplate.Query().Where(imagecreationtemplate.IDEQ(id)).Exist(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return service.ErrImageCreationNotFound
	}
	return service.ErrImageCreationConflict
}

func createImageCreationChangeLog(ctx context.Context, client *dbent.Client, actorID int64, action, targetType, targetID string, metadata map[string]any) error {
	_, err := client.ImageCreationChangeLog.Create().
		SetActorUserID(actorID).
		SetAction(action).
		SetTargetType(targetType).
		SetTargetID(targetID).
		SetMetadata(metadata).
		Save(ctx)
	return err
}

func imageCreationAssetEntityToService(entity *dbent.ImageCreationAsset, withContent bool) *service.ImageCreationAsset {
	if entity == nil {
		return nil
	}
	asset := &service.ImageCreationAsset{
		ID: entity.ID, ContentType: entity.ContentType, ByteSize: entity.ByteSize,
		Width: entity.Width, Height: entity.Height, SourceType: entity.SourceType,
		CreatedBy: entity.CreatedBy, CreatedAt: entity.CreatedAt,
	}
	if withContent {
		asset.Content = entity.Content
	}
	if entity.SourceProvider != nil {
		asset.SourceProvider = *entity.SourceProvider
	}
	if entity.SourceModel != nil {
		asset.SourceModel = *entity.SourceModel
	}
	return asset
}

func imageCreationTemplateEntityToService(entity *dbent.ImageCreationTemplate) *service.ImageCreationTemplate {
	if entity == nil {
		return nil
	}
	return &service.ImageCreationTemplate{
		ID: entity.ID, State: entity.State, DraftData: entity.DraftData, PublishedData: entity.PublishedData,
		Revision: entity.Revision, PublishedVersion: entity.PublishedVersion,
		DraftCoverAssetID: entity.DraftCoverAssetID, PublishedCoverAssetID: entity.PublishedCoverAssetID,
		HomePosition: entity.HomePosition, CreatedBy: entity.CreatedBy, UpdatedBy: entity.UpdatedBy,
		CreatedAt: entity.CreatedAt, UpdatedAt: entity.UpdatedAt, PublishedAt: entity.PublishedAt,
	}
}

func imageCreationTemplateEntitiesToService(entities []*dbent.ImageCreationTemplate) []service.ImageCreationTemplate {
	items := make([]service.ImageCreationTemplate, 0, len(entities))
	for _, entity := range entities {
		items = append(items, *imageCreationTemplateEntityToService(entity))
	}
	return items
}

func imageCreationJSONContainsFold(field string, path string, value string) *entsql.Predicate {
	pattern := "%" + strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(strings.ToLower(value)) + "%"
	return entsql.P(func(b *entsql.Builder) {
		b.WriteString("LOWER(").Ident(field).WriteString("->>'").WriteString(path).WriteString("') LIKE ").Arg(pattern).WriteString(" ESCAPE ").Arg(`\`)
	})
}
