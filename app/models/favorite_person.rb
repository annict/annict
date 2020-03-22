# frozen_string_literal: true
# == Schema Information
#
# Table name: favorite_people
#
#  id                  :bigint           not null, primary key
#  watched_works_count :integer          default("0"), not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#  person_id           :bigint           not null
#  user_id             :bigint           not null
#
# Indexes
#
#  index_favorite_people_on_person_id              (person_id)
#  index_favorite_people_on_user_id                (user_id)
#  index_favorite_people_on_user_id_and_person_id  (user_id,person_id) UNIQUE
#  index_favorite_people_on_watched_works_count    (watched_works_count)
#
# Foreign Keys
#
#  fk_rails_...  (person_id => people.id)
#  fk_rails_...  (user_id => users.id)
#

class FavoritePerson < ApplicationRecord
  belongs_to :person, counter_cache: :favorite_users_count
  belongs_to :user

  scope :with_cast, -> { joins(:person).where("people.casts_count > ?", 0) }
  scope :with_staff, -> { joins(:person).where("people.staffs_count > ?", 0) }

  def update_watched_works_count(user)
    cast_work_ids = person.cast_works.pluck(:id)
    staff_work_ids = person.staff_works.pluck(:id)
    library_entries = user.library_entries.with_not_deleted_work.with_status(:watched)
    count = library_entries.where(work_id: (cast_work_ids | staff_work_ids)).count

    update_column(:watched_works_count, count)
  end
end
