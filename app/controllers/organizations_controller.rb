# == Schema Information
#
# Table name: organizations
#
#  id               :integer          not null, primary key
#  name             :string           not null
#  url              :string
#  wikipedia_url    :string
#  twitter_username :string
#  aasm_state       :string           default("published"), not null
#  created_at       :datetime         not null
#  updated_at       :datetime         not null
#
# Indexes
#
#  index_organizations_on_aasm_state  (aasm_state)
#  index_organizations_on_name        (name) UNIQUE
#

class OrganizationsController < ApplicationController
  def show(id)
    @organization = Organization.published.find(id)
    @wos_with_year = @organization.
      work_organizations.
      includes(work: [:season, :item]).
      group_by { |wo| wo.work.season&.year.presence || 0 }
    @wo_years = @wos_with_year.keys.sort.reverse
  end
end
