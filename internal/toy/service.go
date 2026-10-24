package toy

import (
	"context"
	"strings"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/fc1g/kachenyatko-store-api/internal/category"
	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/fc1g/kachenyatko-store-api/internal/validation"
	"github.com/google/uuid"
)

type toyRepository interface {
	ListToyImagesByToyIDs(ctx context.Context, toyIds []uuid.UUID) ([]sqlc.ToyImage, error)
	CreateToy(ctx context.Context, arg sqlc.CreateToyParams) (sqlc.Toy, error)
	ListToys(ctx context.Context, arg sqlc.ListToysParams) ([]sqlc.ListToysRow, error)
	ListBestsellers(ctx context.Context, arg sqlc.ListBestsellersParams) ([]sqlc.ListBestsellersRow, error)
	CountToys(ctx context.Context, categorySlug *string) (int64, error)
	GetToy(ctx context.Context, id uuid.UUID) (sqlc.GetToyRow, error)
	UpdateToy(ctx context.Context, arg sqlc.UpdateToyParams) (sqlc.Toy, error)
	DeleteToy(ctx context.Context, id uuid.UUID) (int64, error)
}

type toyStore interface {
	toyRepository
	InTx(ctx context.Context, fn func(q sqlc.Querier) error) error
}

type Service struct {
	repo    toyStore
	storage imageStorage
}

func NewService(r toyStore, s imageStorage) *Service {
	return &Service{
		repo:    r,
		storage: s,
	}
}

type Toy struct {
	sqlc.Toy
	Category sqlc.Category
	Images   []sqlc.ToyImage
}

// toyInput is the request body for creating a toy or replacing all of its fields.
type toyInput struct {
	// Category name; the category is created if it does not exist
	CategoryName string `json:"category_name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Soft toys"`
	Name         string `json:"name" validate:"required,max=100" minLength:"1" maxLength:"100" example:"Teddy bear"`
	Description  string `json:"description" validate:"required,max=2000" minLength:"1" maxLength:"2000" example:"Soft plush bear, safe for children from 3 years"`
	// Free-form size, e.g. dimensions in centimeters
	Size      string `json:"size" validate:"required,max=50" minLength:"1" maxLength:"50" example:"25x15 cm"`
	Materials string `json:"materials" validate:"required,max=255" minLength:"1" maxLength:"255" example:"Cotton, polyester filling"`
} // @name ToyInput

func (in *toyInput) normalize() {
	in.CategoryName = strings.TrimSpace(in.CategoryName)
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	in.Size = strings.TrimSpace(in.Size)
	in.Materials = strings.TrimSpace(in.Materials)
}

func (s *Service) Create(ctx context.Context, input toyInput) (Toy, error) {
	input.normalize()

	if err := validation.Struct(input); err != nil {
		return Toy{}, err
	}

	slug, err := category.Slug(input.CategoryName)
	if err != nil {
		return Toy{}, err
	}

	var c sqlc.Category
	var toy sqlc.Toy

	err = s.repo.InTx(ctx, func(q sqlc.Querier) error {
		c, err = q.UpsertCategory(ctx, sqlc.UpsertCategoryParams{
			Name: input.CategoryName,
			Slug: slug,
		})
		if err != nil {
			return categoryError("upsert category", err)
		}

		toy, err = q.CreateToy(ctx, sqlc.CreateToyParams{
			CategoryID:  c.ID,
			Name:        input.Name,
			Description: input.Description,
			Size:        input.Size,
			Materials:   input.Materials,
		})
		if err != nil {
			return toyError("create toy", err)
		}

		return nil
	})
	if err != nil {
		return Toy{}, err
	}

	return Toy{
		Toy:      toy,
		Category: c,
	}, nil
}

type toyPage struct {
	Toys  []Toy
	Total int64
}

func (s *Service) List(ctx context.Context, categorySlug *string, limit, offset int32) (toyPage, error) {
	rows, err := s.repo.ListToys(ctx, sqlc.ListToysParams{
		Limit:        limit,
		Offset:       offset,
		CategorySlug: categorySlug,
	})
	if err != nil {
		return toyPage{}, toyError("get toys", err)
	}

	total, err := s.repo.CountToys(ctx, categorySlug)
	if err != nil {
		return toyPage{}, toyError("count toys", err)
	}

	toys := make([]Toy, 0, len(rows))
	for _, row := range rows {
		toys = append(toys, Toy{Toy: row.Toy, Category: row.Category})
	}

	if err := s.attachImages(ctx, toys); err != nil {
		return toyPage{}, err
	}

	return toyPage{
		Toys:  toys,
		Total: total,
	}, nil
}

func (s *Service) ListBestsellers(ctx context.Context, categorySlug *string, limit, offset int32) (toyPage, error) {
	rows, err := s.repo.ListBestsellers(ctx, sqlc.ListBestsellersParams{
		Limit:        limit,
		Offset:       offset,
		CategorySlug: categorySlug,
	})
	if err != nil {
		return toyPage{}, toyError("get bestsellers", err)
	}

	total, err := s.repo.CountToys(ctx, categorySlug)
	if err != nil {
		return toyPage{}, toyError("count toys", err)
	}

	toys := make([]Toy, 0, len(rows))
	for _, row := range rows {
		toys = append(toys, Toy{Toy: row.Toy, Category: row.Category})
	}

	if err := s.attachImages(ctx, toys); err != nil {
		return toyPage{}, err
	}

	return toyPage{
		Toys:  toys,
		Total: total,
	}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Toy, error) {
	t, err := s.repo.GetToy(ctx, id)
	if err != nil {
		return Toy{}, toyError("get toy", err)
	}

	images, err := s.repo.ListToyImagesByToyIDs(ctx, []uuid.UUID{id})
	if err != nil {
		return Toy{}, imageError("get toy images", err)
	}

	return Toy{
		Toy:      t.Toy,
		Category: t.Category,
		Images:   images,
	}, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input toyInput) (Toy, error) {
	input.normalize()

	if err := validation.Struct(input); err != nil {
		return Toy{}, err
	}

	slug, err := category.Slug(input.CategoryName)
	if err != nil {
		return Toy{}, err
	}

	var c sqlc.Category
	var toy sqlc.Toy
	var images []sqlc.ToyImage

	err = s.repo.InTx(ctx, func(q sqlc.Querier) error {
		c, err = q.UpsertCategory(ctx, sqlc.UpsertCategoryParams{
			Name: input.CategoryName,
			Slug: slug,
		})
		if err != nil {
			return categoryError("upsert category", err)
		}

		toy, err = q.UpdateToy(ctx, sqlc.UpdateToyParams{
			ID:          id,
			CategoryID:  c.ID,
			Name:        input.Name,
			Description: input.Description,
			Size:        input.Size,
			Materials:   input.Materials,
		})
		if err != nil {
			return toyError("update toy", err)
		}

		images, err = q.ListToyImagesByToyIDs(ctx, []uuid.UUID{id})
		if err != nil {
			return imageError("get toy images", err)
		}

		return nil
	})
	if err != nil {
		return Toy{}, err
	}

	return Toy{
		Toy:      toy,
		Category: c,
		Images:   images,
	}, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	var keys []string
	err := s.repo.InTx(ctx, func(q sqlc.Querier) error {
		storageKeys, err := q.DeleteToyImagesByToyID(ctx, id)
		if err != nil {
			return toyError("delete toy images", err)
		}
		keys = storageKeys

		count, err := q.DeleteToy(ctx, id)
		if err != nil {
			return toyError("delete toy", err)
		}

		if count == 0 {
			return apperr.NotFound("toy not found", nil)
		}

		return nil
	})
	if err != nil {
		return err
	}

	deleteImages(ctx, s.storage, keys)

	return nil
}

func (s *Service) attachImages(ctx context.Context, toys []Toy) error {
	if len(toys) == 0 {
		return nil
	}

	ids := make([]uuid.UUID, 0, len(toys))
	for _, t := range toys {
		ids = append(ids, t.ID)
	}

	images, err := s.repo.ListToyImagesByToyIDs(ctx, ids)
	if err != nil {
		return imageError("get toy images", err)
	}

	imagesByToy := make(map[uuid.UUID][]sqlc.ToyImage, len(toys))
	for _, img := range images {
		imagesByToy[img.ToyID] = append(imagesByToy[img.ToyID], img)
	}

	for i := range toys {
		toys[i].Images = imagesByToy[toys[i].ID]
	}

	return nil
}
