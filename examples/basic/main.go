package basic

import "bom/examples/basic/generated"

func main() {
    generated.FindManyVideo(nil, nil, generated.VideoFindMany{})

    generated.FindUniqueVideo(
        nil,
        nil,
        generated.VideoFindUnique[generated.VideoUK_Id]{
            Where: generated.VideoUK_Id{Id: 1},
            Select: generated.VideoSelect{
                generated.VideoFieldAuthorId,
                generated.VideoSelectAuthor{
                    Args: generated.VideoAuthorSelectArgs{
                        Select: generated.AuthorSelect{
                            generated.AuthorSelectVideo{
                                Args: generated.AuthorVideoSelectArgs{
                                    Select: generated.VideoSelectAll,
                                },
                            },
                        },
                    },
                },
            },
        },
    )
}
