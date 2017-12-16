# frozen_string_literal: true
# == Schema Information
#
# Table name: work_taggables
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  work_tag_id :integer          not null
#  description :string
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#  locale      :string           default("other"), not null
#
# Indexes
#
#  index_work_taggables_on_locale                   (locale)
#  index_work_taggables_on_user_id                  (user_id)
#  index_work_taggables_on_user_id_and_work_tag_id  (user_id,work_tag_id) UNIQUE
#  index_work_taggables_on_work_tag_id              (work_tag_id)
#

class WorkTaggable < ApplicationRecord
  belongs_to :user
  belongs_to :work_tag

  validates :description, length: { maximum: 500 }
end
