# frozen_string_literal: true
# == Schema Information
#
# Table name: mute_users
#
#  id              :integer          not null, primary key
#  user_id         :integer          not null
#  ignored_user_id :integer          not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_mute_users_on_ignored_user_id  (ignored_user_id)
#  index_mute_users_on_user_id          (user_id)
#


class MuteUser < ApplicationRecord
end
