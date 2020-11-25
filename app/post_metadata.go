package app

import (
	"github.com/masterhung0112/hk_server/model"
)

func (a *App) PreparePostForClient(originalPost *model.Post, isNewPost bool, isEditPost bool) *model.Post {
	post := originalPost.Clone()

	//TODO: Open
	// Proxy image links before constructing metadata so that requests go through the proxy
	// post = a.PostWithProxyAddedToImageURLs(post)

	// a.OverrideIconURLIfEmoji(post)

	// post.Metadata = &model.PostMetadata{}

	// if post.DeleteAt > 0 {
	// 	// For deleted posts we don't fill out metadata nor do we return the post content
	// 	post.Message = ""
	// 	return post
	// }

	// // Emojis and reaction counts
	// if emojis, reactions, err := a.getEmojisAndReactionsForPost(post); err != nil {
	// 	mlog.Warn("Failed to get emojis and reactions for a post", mlog.String("post_id", post.Id), mlog.Err(err))
	// } else {
	// 	post.Metadata.Emojis = emojis
	// 	post.Metadata.Reactions = reactions
	// }

	// // Files
	// if fileInfos, err := a.getFileMetadataForPost(post, isNewPost || isEditPost); err != nil {
	// 	mlog.Warn("Failed to get files for a post", mlog.String("post_id", post.Id), mlog.Err(err))
	// } else {
	// 	post.Metadata.Files = fileInfos
	// }

	// // Embeds and image dimensions
	// firstLink, images := getFirstLinkAndImages(post.Message)

	// if embed, err := a.getEmbedForPost(post, firstLink, isNewPost); err != nil {
	// 	mlog.Debug("Failed to get embedded content for a post", mlog.String("post_id", post.Id), mlog.Err(err))
	// } else if embed == nil {
	// 	post.Metadata.Embeds = []*model.PostEmbed{}
	// } else {
	// 	post.Metadata.Embeds = []*model.PostEmbed{embed}
	// }

	// post.Metadata.Images = a.getImagesForPost(post, images, isNewPost)

	return post
}
