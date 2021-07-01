# frozen_string_literal: true

class FavoritePeopleController < ApplicationV6Controller
  def index
    set_page_category PageCategory::FAVORITE_PERSON_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @cast_favorites = @user
      .person_favorites
      .preload(:person)
      .with_cast
      .order(watched_works_count: :desc)
    @staff_favorites = @user
      .person_favorites
      .preload(:person)
      .with_staff
      .order(watched_works_count: :desc)
  end
end
