# frozen_string_literal: true

class OrganizationFansController < ApplicationV6Controller
  def index
    set_page_category PageCategory::ORGANIZATION_FAN_LIST

    @organization = Organization.only_kept.find(params[:organization_id])
    @organization_favorites = @organization
      .organization_favorites
      .eager_load(user: :profile)
      .merge(User.only_kept)
      .order(watched_works_count: :desc)
  end
end
