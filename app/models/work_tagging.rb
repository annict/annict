# frozen_string_literal: true
# == Schema Information
#
# Table name: work_taggings
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  work_id     :integer          not null
#  work_tag_id :integer          not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#
# Indexes
#
#  index_work_taggings_on_user_id                              (user_id)
#  index_work_taggings_on_user_id_and_work_id_and_work_tag_id  (user_id,work_id,work_tag_id) UNIQUE
#  index_work_taggings_on_work_id                              (work_id)
#  index_work_taggings_on_work_tag_id                          (work_tag_id)
#

class WorkTagging < ApplicationRecord
  belongs_to :user
  belongs_to :work
  belongs_to :work_tag, counter_cache: true
end
