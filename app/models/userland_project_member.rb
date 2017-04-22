# frozen_string_literal: true
# == Schema Information
#
# Table name: userland_project_members
#
#  id                  :integer          not null, primary key
#  user_id             :integer          not null
#  userland_project_id :integer          not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#
# Indexes
#
#  index_userland_pm_on_uid_and_userland_pid              (user_id,userland_project_id) UNIQUE
#  index_userland_project_members_on_user_id              (user_id)
#  index_userland_project_members_on_userland_project_id  (userland_project_id)
#

class UserlandProjectMember < ApplicationRecord
  belongs_to :user
  belongs_to :userland_project
end
