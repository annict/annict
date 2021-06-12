# frozen_string_literal: true

class OrganizationsController < ApplicationController
  before_action :load_i18n, only: %i[show]

  def show
    @organization = Organization.only_kept.find(params[:organization_id])
    @staffs_with_year = @organization
      .staffs
      .only_kept
      .joins(:work)
      .where(works: {deleted_at: nil})
      .includes(work: :work_image)
      .group_by { |s| s.work.season_year.presence || 0 }
    @staff_years = @staffs_with_year.keys.sort.reverse

    @organization_favorites = @organization
      .organization_favorites
      .eager_load(user: :profile)
      .merge(User.only_kept)
      .order(id: :desc)
  end

  private

  def load_i18n
    keys = {
      "messages._components.favorite_button.add_to_favorites": nil,
      "messages._components.favorite_button.added_to_favorites": nil
    }

    load_i18n_into_gon keys
  end
end
