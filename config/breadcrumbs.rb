# frozen_string_literal: true

crumb :root do
  link t("noun.home"), root_path
end

crumb :collection_list do
  link t("noun.collections"), collections_path
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

crumb :work_detail do |work|
  link work.decorate.local_title, work_path(work)
  parent :root
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

crumb :user_collection_list do |user|
  link t("noun.collection_list"), user_collections_path(user.username)
  parent :user_detail, user
end

crumb :user_collection_detail do |collection|
  user = collection.user
  link collection.title, user_collection_path(user.username, collection)
  parent :user_collection_list, user
end

crumb :edit_collection_item do |collection_item|
  link collection_item.title, edit_collection_collection_item_path(collection_item.collection, collection_item)
  parent :user_collection_detail, collection_item.collection
end

crumb :edit_collection do |collection|
  link collection.title, edit_collection_path(collection)
  parent :user_collection_detail, collection
end
