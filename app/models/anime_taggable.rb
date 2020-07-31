# frozen_string_literal: true
# == Schema Information
#
# Table name: anime_taggables
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
#  index_anime_taggables_on_locale                   (locale)
#  index_anime_taggables_on_user_id                  (user_id)
#  index_anime_taggables_on_user_id_and_work_tag_id  (user_id,work_tag_id) UNIQUE
#  index_anime_taggables_on_work_tag_id              (work_tag_id)
#
# Foreign Keys
#
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_tag_id => anime_tags.id)
#

class AnimeTaggable < ApplicationRecord
  belongs_to :user
  belongs_to :work_tag

  validates :description, length: { maximum: 500 }
end
