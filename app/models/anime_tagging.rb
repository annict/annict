# frozen_string_literal: true
# == Schema Information
#
# Table name: anime_taggings
#
#  id          :bigint           not null, primary key
#  created_at  :datetime         not null
#  updated_at  :datetime         not null
#  anime_id    :bigint           not null
#  user_id     :bigint           not null
#  work_tag_id :bigint           not null
#
# Indexes
#
#  index_anime_taggings_on_anime_id                              (anime_id)
#  index_anime_taggings_on_user_id                               (user_id)
#  index_anime_taggings_on_user_id_and_anime_id_and_work_tag_id  (user_id,anime_id,work_tag_id) UNIQUE
#  index_anime_taggings_on_work_tag_id                           (work_tag_id)
#
# Foreign Keys
#
#  fk_rails_...  (anime_id => animes.id)
#  fk_rails_...  (user_id => users.id)
#  fk_rails_...  (work_tag_id => anime_tags.id)
#

class AnimeTagging < ApplicationRecord
  counter_culture :work_tag

  belongs_to :user
  belongs_to :work
  belongs_to :work_tag
end
