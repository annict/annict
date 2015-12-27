# == Schema Information
#
# Table name: work_organizations
#
#  id              :integer          not null, primary key
#  work_id         :integer          not null
#  organization_id :integer          not null
#  role            :string           not null
#  role_other      :string
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_work_organizations_on_organization_id              (organization_id)
#  index_work_organizations_on_work_id                      (work_id)
#  index_work_organizations_on_work_id_and_organization_id  (work_id,organization_id) UNIQUE
#

class WorkOrganization < ActiveRecord::Base
  extend Enumerize

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  enumerize :role, in: %w(
    producer
  )

  belongs_to :organization
  belongs_to :work
end
