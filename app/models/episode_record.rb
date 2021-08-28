# frozen_string_literal: true

# == Schema Information
#
# Table name: episode_records
#
#  id                :bigint           not null, primary key
#  facebook_url_hash :string(510)
#  twitter_url_hash  :string(510)
#  created_at        :datetime
#  updated_at        :datetime
#
# Indexes
#
#  checkins_facebook_url_hash_key                       (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key                        (twitter_url_hash) UNIQUE
#  checkins_user_id_idx                                 (user_id)
#  index_episode_records_on_episode_id_and_deleted_at   (episode_id,deleted_at)
#  index_episode_records_on_locale                      (locale)
#  index_episode_records_on_multiple_episode_record_id  (multiple_episode_record_id)
#  index_episode_records_on_oauth_application_id        (oauth_application_id)
#  index_episode_records_on_rating_state                (rating_state)
#  index_episode_records_on_record_id                   (record_id) UNIQUE
#  index_episode_records_on_review_id                   (review_id)
#  index_episode_records_on_work_id                     (work_id)
#
# Foreign Keys
#
#  checkins_episode_id_fk  (episode_id => episodes.id) ON DELETE => cascade
#  checkins_user_id_fk     (user_id => users.id) ON DELETE => cascade
#  checkins_work_id_fk     (work_id => works.id)
#  fk_rails_...            (multiple_episode_record_id => multiple_episode_records.id)
#  fk_rails_...            (oauth_application_id => oauth_applications.id)
#  fk_rails_...            (record_id => records.id)
#  fk_rails_...            (review_id => work_records.id)
#

class EpisodeRecord < ApplicationRecord
  include Recordable

  self.ignored_columns = %w[
    aasm_state
    body
    comments_count
    deleted_at
    episode_id
    facebook_click_count
    likes_count
    locale
    modify_body
    multiple_episode_record_id
    oauth_application_id
    rating
    rating_state
    record_id
    review_id
    shared_facebook
    shared_twitter
    twitter_click_count
    user_id
    work_id
  ]
end
