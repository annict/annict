# frozen_string_literal: true
# == Schema Information
#
# Table name: user_work_comments
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  work_id    :integer          not null
#  body       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_user_work_comments_on_user_id              (user_id)
#  index_user_work_comments_on_user_id_and_work_id  (user_id,work_id) UNIQUE
#  index_user_work_comments_on_work_id              (work_id)
#

class UserWorkComment < ApplicationRecord
  belongs_to :user
  belongs_to :work

  # validates :body, presence: true, length: { maximum: 150 }
end
