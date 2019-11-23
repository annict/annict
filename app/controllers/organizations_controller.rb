# frozen_string_literal: true

class OrganizationsController < ApplicationController
  before_action :load_i18n, only: %i(show)

  def show
    @organization = Organization.without_deleted.find(params[:id])
    @staffs_with_year = @organization.
      staffs.
      without_deleted.
      joins(:work).
      where(works: { deleted_at: nil }).
      includes(work: :work_image).
      group_by { |s| s.work.season_year.presence || 0 }
    @staff_years = @staffs_with_year.keys.sort.reverse

    @favorite_orgs = @organization.
      favorite_organizations.
      joins(:user).
      merge(User.without_deleted).
      order(id: :desc)
  end

  private

  def load_i18n
    keys = {
      "messages._components.favorite_button.add_to_favorites": nil,
      "messages._components.favorite_button.added_to_favorites": nil,
    }

    load_i18n_into_gon keys
  end
end
