# frozen_string_literal: true
# == Schema Information
#
# Table name: checkins
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
#  checkins_episode_id_idx                 (episode_id)
#  checkins_facebook_url_hash_key          (facebook_url_hash) UNIQUE
#  checkins_twitter_url_hash_key           (twitter_url_hash) UNIQUE
#  checkins_user_id_idx                    (user_id)
#  index_checkins_on_multiple_record_id    (multiple_record_id)
#  index_checkins_on_oauth_application_id  (oauth_application_id)
#  index_checkins_on_work_id               (work_id)
#

class CheckinsController < ApplicationController
  def redirect(provider, url_hash)
    case provider
    when "tw"
      checkin = Checkin.find_by!(twitter_url_hash: url_hash)

      log_messages = [
        "Twitterからのアクセス ",
        "remote_host: #{request.remote_host}, ",
        "remote_ip: #{request.remote_ip}, ",
        "remote_user: #{request.remote_user}"
      ]
      logger.info(log_messages.join)

      bots = TwitterBot.pluck(:name)
      no_bots = bots.map do |bot|
        request.user_agent.present? && !request.user_agent.include?(bot)
      end
      checkin.increment!(:twitter_click_count) if no_bots.all?

      redirect_to_user_record(checkin)
    when "fb"
      checkin = Checkin.find_by!(facebook_url_hash: url_hash)
      checkin.increment!(:facebook_click_count)

      redirect_to_user_record(checkin)
    else
      redirect_to root_path
    end
  end

  private

  def redirect_to_user_record(checkin)
    username = checkin.user.username

    redirect_to record_path(username, checkin)
  end
end
