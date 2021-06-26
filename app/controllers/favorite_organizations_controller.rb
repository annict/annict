# frozen_string_literal: true

class FavoriteOrganizationsController < ApplicationV6Controller
  def index
    set_page_category PageCategory::FAVORITE_ORGANIZATION_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @organization_favorites = @user
      .organization_favorites
      .preload(:organization)
      .order(watched_works_count: :desc)
  end
end
