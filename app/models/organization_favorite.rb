# typed: false
# frozen_string_literal: true

class OrganizationFavorite < ApplicationRecord
  counter_culture :organization, column_name: :favorite_users_count
  counter_culture :user

  belongs_to :organization
  belongs_to :user

  def watched_work_count
    watched_works_count
  end

  def update_watched_works_count(user)
    staff_work_ids = organization.staff_works.pluck(:id)
    library_entries = user.library_entries.with_not_deleted_work.with_status(:watched)
    count = library_entries.where(work_id: staff_work_ids).count

    update_column(:watched_works_count, count)
  end
end
