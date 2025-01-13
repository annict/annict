# typed: false
# frozen_string_literal: true

class PersonFavorite < ApplicationRecord
  counter_culture :person, column_name: :favorite_users_count
  counter_culture :user

  belongs_to :person
  belongs_to :user

  scope :with_cast, -> { joins(:person).where("people.casts_count > ?", 0) }
  scope :with_staff, -> { joins(:person).where("people.staffs_count > ?", 0) }

  def watched_work_count
    watched_works_count
  end

  def update_watched_works_count(user)
    cast_work_ids = person.cast_works.pluck(:id)
    staff_work_ids = person.staff_works.pluck(:id)
    library_entries = user.library_entries.with_not_deleted_work.with_status(:watched)
    count = library_entries.where(work_id: (cast_work_ids | staff_work_ids)).count

    update_column(:watched_works_count, count)
  end
end
