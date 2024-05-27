# typed: false
# frozen_string_literal: true

# == Schema Information
#
# Table name: mute_users
#
#  id            :bigint           not null, primary key
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#  muted_user_id :bigint           not null
#  user_id       :bigint           not null
#
# Indexes
#
#  index_mute_users_on_muted_user_id              (muted_user_id)
#  index_mute_users_on_user_id                    (user_id)
#  index_mute_users_on_user_id_and_muted_user_id  (user_id,muted_user_id) UNIQUE
#
# Foreign Keys
#
#  fk_rails_...  (muted_user_id => users.id)
#  fk_rails_...  (user_id => users.id)
#

class MuteUser < ApplicationRecord
  belongs_to :user
  belongs_to :muted_user, class_name: "User"
end
