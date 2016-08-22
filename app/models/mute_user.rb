# frozen_string_literal: true
# == Schema Information
#
# Table name: mute_users
#
#  id              :integer          not null, primary key
#  user_id         :integer          not null
#  muted_user_id :integer          not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_mute_users_on_muted_user_id              (muted_user_id)
#  index_mute_users_on_user_id                      (user_id)
#  index_mute_users_on_user_id_and_muted_user_id  (user_id,muted_user_id) UNIQUE
#

class MuteUser < ApplicationRecord
  belongs_to :user
  belongs_to :muted_user, class_name: "User"
end
