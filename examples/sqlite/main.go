package sqlite

import (
	"context"

	"github.com/10fu3/bom/examples/sqlite/generated"
	"github.com/10fu3/bom/pkg/bom"
	"github.com/10fu3/bom/pkg/opt"
)

// main exists so `go test ./...` can build the example module.
func main() {}

func sampleHasOne(ctx context.Context, db bom.Querier) {
	_, _ = generated.FindManyAuthor[generated.Author](ctx, db, generated.AuthorFindMany{
		Select: generated.AuthorSelect{
			generated.AuthorFieldName,
			generated.AuthorSelectAuthorProfile{
				Args: generated.AuthorAuthorProfileSelectArgs{
					Select: generated.AuthorProfileSelect{
						generated.AuthorProfileFieldAvatarUrl,
					},
				},
			},
		},
	})
}

func sampleBelongsTo(ctx context.Context, db bom.Querier) {
	_, _ = generated.FindManyVideo[generated.Video](ctx, db, generated.VideoFindMany{
		Select: generated.VideoSelect{
			generated.VideoFieldTitle,
			generated.VideoSelectAuthor{
				Args: generated.VideoAuthorSelectArgs{
					Select: generated.AuthorSelect{
						generated.AuthorFieldName,
					},
				},
			},
		},
	})
}

func sampleHasMany(ctx context.Context, db bom.Querier) {
	_, _ = generated.FindManyAuthor[generated.Author](ctx, db, generated.AuthorFindMany{
		Select: generated.AuthorSelect{
			generated.AuthorFieldName,
			generated.AuthorSelectVideo{
				Args: generated.AuthorVideoSelectArgs{
					Select: generated.VideoSelect{
						generated.VideoFieldTitle,
					},
				},
			},
		},
	})
}

func sampleRelationFilters(ctx context.Context, db bom.Querier) {
	_, _ = generated.FindManyAuthor[generated.Author](ctx, db, generated.AuthorFindMany{
		Where: &generated.AuthorWhereInput{
			Video: &generated.AuthorVideoRelation{
				Some: &generated.VideoWhereInput{
					Comment: &generated.VideoCommentRelation{
						Some: &generated.CommentWhereInput{
							Body: opt.OVal("great"),
						},
					},
				},
				None: &generated.VideoWhereInput{
					Comment: &generated.VideoCommentRelation{
						Some: &generated.CommentWhereInput{
							Body: opt.OVal("spam"),
						},
					},
				},
				Every: &generated.VideoWhereInput{
					Comment: &generated.VideoCommentRelation{
						None: &generated.CommentWhereInput{
							Body: opt.OVal("spam"),
						},
					},
				},
			},
		},
		Select: generated.AuthorSelect{
			generated.AuthorFieldName,
		},
	})
}
