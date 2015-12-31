# == Schema Information
#
# Table name: work_organizations
#
#  id              :integer          not null, primary key
#  work_id         :integer          not null
#  organization_id :integer          not null
#  role            :string           not null
#  role_other      :string
#  aasm_state      :string           default("published"), not null
#  sort_number     :integer          default(0), not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_work_organizations_on_aasm_state                   (aasm_state)
#  index_work_organizations_on_organization_id              (organization_id)
#  index_work_organizations_on_sort_number                  (sort_number)
#  index_work_organizations_on_work_id                      (work_id)
#  index_work_organizations_on_work_id_and_organization_id  (work_id,organization_id) UNIQUE
#

class WorkOrganization < ActiveRecord::Base
  include AASM
  include DbActivityMethods
  include WorkOrganizationCommon

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end
end
