# frozen_string_literal: true
# == Schema Information
#
# Table name: twitter_users
#
#  id          :bigint(8)        not null, primary key
#  work_id     :integer
#  screen_name :string           not null
#  user_id     :string
#  aasm_state  :string           default("published"), not null
#  followed_at :datetime
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_twitter_users_on_screen_name  (screen_name) UNIQUE
#  index_twitter_users_on_user_id      (user_id) UNIQUE
#  index_twitter_users_on_work_id      (work_id)
#

class TwitterUser < ApplicationRecord
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  belongs_to :work
end
