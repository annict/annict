# frozen_string_literal: true

# == Schema Information
#
# Table name: person_favorites
#
#  id                  :bigint           not null, primary key
#  watched_works_count :integer          default(0), not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#  person_id           :bigint           not null
#  user_id             :bigint           not null
#
# Indexes
#
#  index_person_favorites_on_person_id              (person_id)
#  index_person_favorites_on_user_id                (user_id)
#  index_person_favorites_on_user_id_and_person_id  (user_id,person_id) UNIQUE
#  index_person_favorites_on_watched_works_count    (watched_works_count)
#
# Foreign Keys
#
#  fk_rails_...  (person_id => people.id)
#  fk_rails_...  (user_id => users.id)
#

class PersonFavorite < ApplicationRecord
  counter_culture :person, column_name: :favorite_users_count
  counter_culture :user

  belongs_to :person
  belongs_to :user

  scope :with_cast, -> { joins(:person).where("people.casts_count > ?", 0) }
  scope :with_staff, -> { joins(:person).where("people.staffs_count > ?", 0) }

  def watched_anime_count
    watched_works_count
  end

  def update_watched_works_count(user)
    cast_work_ids = person.cast_works.pluck(:id)
    staff_work_ids = person.staff_works.pluck(:id)
    library_entries = user.library_entries.with_not_deleted_anime.with_status(:watched)
    count = library_entries.where(work_id: (cast_work_ids | staff_work_ids)).count

    update_column(:watched_works_count, count)
  end
end
