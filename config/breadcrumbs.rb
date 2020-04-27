# frozen_string_literal: true

crumb :root do
  link t("noun.home"), root_path
end

crumb :edit_episode_record do |episode_record|
  episode = episode_record.episode
  link_title = I18n.t("head.title.episode_records.edit")
  link link_title, edit_episode_record_path(episode, episode_record)
  parent :episode_detail, episode
end

crumb :edit_work_record do |work_record|
  link t("head.title.work_records.edit"), edit_work_record_path(work_record.work, work_record.record)
  parent :work_record_detail, work_record
end

crumb :episode_detail do |episode|
  link episode.title_with_number(fallback: false), work_episode_path(episode.work, episode)
  parent :work_detail, episode.work
end


crumb :episode_record_detail do |episode_record|
  user = episode_record.user
  work = episode_record.work
  episode = episode_record.episode
  link_title = I18n.t("noun.record_of_episode", work_title: work.local_title, episode_title_with_number: episode.title_with_number)
  link link_title, record_path(user.username, episode_record.record)
  parent :user_record_list, user
end

crumb :faq_list do
  link t("noun.faqs"), faqs_path
  parent :root
end

crumb :forum_root do
  link "Forum", forum_root_path
  parent :root
end

crumb :forum_category_detail do |category|
  link category.local_name, forum_category_path(category.slug)
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
  link post.title.truncate(10), forum_post_path(post)
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

crumb :friends_index do
  link t("head.title.friends.index")
  parent :root
end

crumb :newest_works do
  link t("head.title.works.newest")
  parent :root
end

crumb :notifications_index do
  link t("head.title.notifications.index")
  parent :root
end

crumb :pages_legal do
  link t("head.title.pages.legal")
  parent :root
end

crumb :popular_works do
  link t("head.title.works.popular")
  parent :root
end

crumb :slots_index do
  link t("head.title.slots.index")
  parent :root
end

crumb :seasonal_works do |season_slug, season_name|
  link t("head.title.works.season", name: season_name), season_works_path(slug: season_slug)
  parent :root
end

crumb :supporters_index do
  link t("head.title.supporters.index"), supporters_path
  parent :root
end

crumb :user_detail do |user|
  link user.profile.name, user_path(user.username)
  parent :root
end

crumb :user_record_list do |user|
  link t("noun.record_list"), records_path(user.username)
  parent :user_detail, user
end

crumb :user_work_tag_detail do |user, tag|
  link tag.name, user_work_tag_path(user.username, tag.name)
  parent :user_work_tag_list, user
end

crumb :user_work_tag_list do |user|
  link t("noun.tags")
  parent :user_detail, user
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

crumb :work_detail do |work|
  link work.local_title, work_path(work)

  if work.season.present?
    parent :seasonal_works, work.season.slug, work.season.local_name
  else
    parent :root
  end
end

crumb :work_detail_v4 do |work|
  link work.local_title, work_path(work)

  if work.season_slug.present?
    parent :seasonal_works, work.season_slug, work.local_season_name
  else
    parent :root
  end
end

crumb :work_episode_list do |work|
  link t("noun.episodes"), work_episodes_path(work)
  parent :work_detail, work
end

crumb :work_record_detail do |work_record|
  user = work_record.user
  work = work_record.work
  link_title = t("noun.record_of_work", work_title: work.local_title)
  link link_title, record_path(user.username, work_record.record)
  parent :user_record_list, user
end

crumb :work_record_list do |work|
  link t("noun.records"), work_records_path(work)
  parent :work_detail, work
end
