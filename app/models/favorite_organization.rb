# frozen_string_literal: true
# == Schema Information
#
# Table name: favorite_organizations
#
#  id                  :integer          not null, primary key
#  watched_works_count :integer          default(0), not null
#  created_at          :datetime         not null
#  updated_at          :datetime         not null
#  organization_id     :integer          not null
#  user_id             :integer          not null
#
# Indexes
#
#  index_favorite_organizations_on_organization_id              (organization_id)
#  index_favorite_organizations_on_user_id                      (user_id)
#  index_favorite_organizations_on_user_id_and_organization_id  (user_id,organization_id) UNIQUE
#  index_favorite_organizations_on_watched_works_count          (watched_works_count)
#
# Foreign Keys
#
#  fk_rails_...  (organization_id => organizations.id)
#  fk_rails_...  (user_id => users.id)
#

class FavoriteOrganization < ApplicationRecord
  belongs_to :organization
  belongs_to :user, counter_cache: true

  def update_watched_works_count(user)
    staff_work_ids = organization.staff_works.pluck(:id)
    statuses = user.statuses.work_published.with_kind(:watched)
    count = statuses.where(work_id: staff_work_ids).count

    update_column(:watched_works_count, count)
  end
end
