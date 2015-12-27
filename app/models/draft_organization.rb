# == Schema Information
#
# Table name: draft_organizations
#
#  id               :integer          not null, primary key
#  organization_id  :integer
#  name             :string           not null
#  url              :string
#  wikipedia_url    :string
#  twitter_username :string
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#
# Indexes
#
#  index_draft_organizations_on_name             (name)
#  index_draft_organizations_on_organization_id  (organization_id)
#

class DraftOrganization < ActiveRecord::Base
  include DraftCommon
  include OrganizationCommon

  belongs_to :origin, class_name: "Organization", foreign_key: :organization_id
end
