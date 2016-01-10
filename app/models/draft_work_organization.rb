# == Schema Information
#
# Table name: draft_work_organizations
#
#  id                   :integer          not null, primary key
#  work_organization_id :integer
#  work_id              :integer          not null
#  organization_id      :integer          not null
#  role                 :string           not null
#  role_other           :string
#  sort_number          :integer          default(0), not null
#  created_at           :datetime         not null
#  updated_at           :datetime         not null
#
# Indexes
#
#  index_draft_work_organizations_on_organization_id       (organization_id)
#  index_draft_work_organizations_on_sort_number           (sort_number)
#  index_draft_work_organizations_on_work_id               (work_id)
#  index_draft_work_organizations_on_work_organization_id  (work_organization_id)
#

class DraftWorkOrganization < ActiveRecord::Base
  include DraftCommon
  include WorkOrganizationCommon

  belongs_to :origin, class_name: "WorkOrganization", foreign_key: :work_organization_id
end
