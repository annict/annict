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
