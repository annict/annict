# frozen_string_literal: true
# == Schema Information
#
# Table name: favorite_people
#
#  id                  :integer          not null, primary key
#  user_id             :integer          not null
#  person_id           :integer          not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#  watched_works_count :integer          default(0), not null
#
# Indexes
#
#  index_favorite_people_on_person_id              (person_id)
#  index_favorite_people_on_user_id                (user_id)
#  index_favorite_people_on_user_id_and_person_id  (user_id,person_id) UNIQUE
#  index_favorite_people_on_watched_works_count    (watched_works_count)
#

class FavoritePerson < ApplicationRecord
  belongs_to :person, counter_cache: true
  belongs_to :user

  scope :with_cast, -> { joins(:person).where("people.casts_count > ?", 0) }
  scope :with_staff, -> { joins(:person).where("people.staffs_count > ?", 0) }
end
