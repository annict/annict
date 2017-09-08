# frozen_string_literal: true

crumb :root do
  link t("noun.home"), root_path
end

crumb :faq_list do
  link t("noun.faqs"), faqs_path
  parent :root
end

crumb :userland_root do
  link "Userland", userland_root_path
  parent :root
end

crumb :userland_new_project do
  link t("resources.userland_project.new"), new_userland_project_path
  parent :userland_root
end

crumb :userland_project_detail do |project|
  link project.name, userland_project_path(project)
  parent :userland_root
end

crumb :userland_edit_project do |project|
  link t("resources.userland_project.edit"), edit_userland_project_path(project)
  parent :userland_project_detail, project
end

crumb :user_detail do |user|
  link user.profile.name, user_path(user.username)
  parent :root
end

crumb :seasonal_works do |season_slug, season_name|
  link t("head.title.works.season", name: season_name), season_works_path(slug: season_slug)
  parent :root
end

crumb :work_detail do |work|
  link work.decorate.local_title, work_path(work)

  if work.season.present?
    parent :seasonal_works, work.season.slug, work.season.local_name
  else
    parent :root
  end
end

crumb :user_review_list do |user|
  link t("noun.review_list"), reviews_path(user.username)
  parent :user_detail, user
end

crumb :work_review_list do |work|
  link t("noun.reviews"), work_reviews_path(work)
  parent :work_detail, work
end

crumb :new_review do |work|
  link t("head.title.reviews.new"), new_work_review_path(work)
  parent :work_review_list, work
end

crumb :edit_review do |review|
  link t("head.title.reviews.edit"), edit_work_review_path(review.work, review)
  parent :review_detail, review
end

crumb :review_detail do |review|
  user = review.user
  work = review.work
  link_title = t("noun.review_of_the_work", work_title: work.decorate.local_title)
  link link_title, review_path(user.username, review)
  parent :user_review_list, user
end

crumb :work_character_list do |work|
  link t("noun.characters"), work_characters_path(work)
  parent :work_detail, work
end

crumb :work_staff_list do |work|
  link t("noun.staffs"), work_staffs_path(work)
  parent :work_detail, work
end

crumb :work_episode_list do |work|
  link t("noun.episodes"), work_episodes_path(work)
  parent :work_detail, work
end

crumb :new_work_item do |work|
  link t("head.title.work_items.new"), new_work_item_path(work)
  parent :work_item_list, work
end

crumb :new_episode_item do |episode|
  link t("head.title.episode_items.new"), new_episode_item_path(episode)
  parent :episode_detail, episode
end

crumb :work_item_list do |work|
  link t("noun.related_items"), work_items_path(work)
  parent :work_detail, work
end

crumb :episode_detail do |episode|
  link episode.decorate.title_with_number, work_episode_path(episode.work, episode)
  parent :work_detail, episode.work
end

crumb :user_work_tag_list do |user|
  link t("noun.tags")
  parent :user_detail, user
end

crumb :user_work_tag_detail do |user, tag|
  link tag.name, user_work_tag_path(user.username, tag.name)
  parent :user_work_tag_list, user
end

crumb :forum_root do
  link "Forum", forum_root_path
  parent :root
end

crumb :forum_category_detail do |category|
  link category.decorate.local_name, forum_category_path(category.slug)
  parent :forum_root
end

crumb :forum_new_post do |post|
  link t("head.title.forum.posts.new")

  if post.forum_category.present?
    parent :forum_category_detail, post.forum_category
  else
    parent :forum_root
  end
end

crumb :forum_post_detail do |post|
  link post.title, forum_post_path(post)
  parent :forum_category_detail, post.forum_category
end

crumb :forum_edit_post do |post|
  link t("head.title.forum.posts.edit"), edit_forum_post_path(post)
  parent :forum_post_detail, post
end

crumb :forum_edit_comment do |comment|
  link t("head.title.forum.comments.edit"), edit_forum_post_comment_path(comment.forum_post, comment)
  parent :forum_post_detail, comment.forum_post
end
