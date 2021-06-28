# frozen_string_literal: true

# == Schema Information
#
# Table name: work_taggables
#
#  id          :bigint           not null, primary key
#  description :string
#  locale      :string           default("other"), not null
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#  user_id     :bigint           not null
#  work_tag_id :bigint           not null
#
# Indexes
#
#  index_work_taggables_on_locale                   (locale)
#  index_work_taggables_on_user_id                  (user_id)
#  index_work_taggables_on_user_id_and_work_tag_id  (user_id,work_tag_id) UNIQUE
#  index_work_taggables_on_work_tag_id              (work_tag_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_tag_id => work_tags.id)
#

class AnimeTaggable < ApplicationRecord
  self.table_name = "work_taggables"

  belongs_to :user
  belongs_to :anime_tag

  validates :description, length: {maximum: 500}
end
