# frozen_string_literal: true
# == Schema Information
#
# Table name: work_comments
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  work_id    :integer          not null
#  body       :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#  locale     :string           default("other"), not null
#
# Indexes
#
#  index_work_comments_on_locale               (locale)
#  index_work_comments_on_user_id              (user_id)
#  index_work_comments_on_user_id_and_work_id  (user_id,work_id) UNIQUE
#  index_work_comments_on_work_id              (work_id)
#

class WorkComment < ApplicationRecord
  belongs_to :user
  belongs_to :work

  validates :body, length: { maximum: 150 }, allow_blank: true
end
