# frozen_string_literal: true
# == Schema Information
#
# Table name: email_notifications
#
#  id                          :integer          not null, primary key
#  user_id                     :integer          not null
#  unsubscription_key          :string           not null
#  event_followed_user         :boolean          default(TRUE), not null
#  event_liked_record          :boolean          default(TRUE), not null
#  event_liked_multiple_record :boolean          default(TRUE), not null
#  event_liked_comment         :boolean          default(TRUE), not null
#  event_liked_status          :boolean          default(TRUE), not null
#  event_commented             :boolean          default(TRUE), not null
#  event_friend_joined         :boolean          default(TRUE), not null
#  event_next_season_came      :boolean          default(TRUE), not null
#  created_at                  :datetime         not null
#  updated_at                  :datetime         not null
#
# Indexes
#
#  index_email_notifications_on_unsubscription_key  (unsubscription_key) UNIQUE
#  index_email_notifications_on_user_id             (user_id) UNIQUE
#


class EmailNotification < ApplicationRecord
  EVENTS = %w(
    followed
    liked
    commented
    friend_joined
    next_season_came
  ).freeze
end
