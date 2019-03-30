# frozen_string_literal: true
# == Schema Information
#
# Table name: records
#
#  id                   :integer          not null, primary key
#  user_id              :integer          not null
#  episode_id           :integer          not null
#  comment              :text
#  modify_comment       :boolean          default(FALSE), not null
#  twitter_url_hash     :string(510)
#  facebook_url_hash    :string(510)
#  twitter_click_count  :integer          default(0), not null
#  facebook_click_count :integer          default(0), not null
#  comments_count       :integer          default(0), not null
#  likes_count          :integer          default(0), not null
#  created_at           :datetime
#  updated_at           :datetime
#  shared_twitter       :boolean          default(FALSE), not null
#  shared_facebook      :boolean          default(FALSE), not null
#  work_id              :integer          not null
#  rating               :float
#  multiple_record_id   :integer
#  oauth_application_id :integer
#
# Indexes
#
#  records_episode_id_idx                 (episode_id)
#  records_facebook_url_hash_key          (facebook_url_hash) UNIQUE
#  records_twitter_url_hash_key           (twitter_url_hash) UNIQUE
#  records_user_id_idx                    (user_id)
#  index_records_on_multiple_record_id    (multiple_record_id)
#  index_records_on_oauth_application_id  (oauth_application_id)
#  index_records_on_work_id               (work_id)
#

class CheckinsController < ApplicationController
  # Old record page
  def show
    record = Record.published.find(params[:id])
    redirect_to record_path(record.user.username, record), status: 301
  end
end
