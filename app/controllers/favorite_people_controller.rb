# frozen_string_literal: true

class FavoritePeopleController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index
    set_page_category PageCategory::FAVORITE_PERSON_LIST

    @user = User.only_kept.find_by!(username: params[:username])
    @user_entity = UserEntity.from_model(@user)
    @cast_favorites = @user.
      person_favorites.
      preload(:person).
      with_cast.
      order(watched_works_count: :desc)
    @staff_favorites = @user.
      person_favorites.
      preload(:person).
      with_staff.
      order(watched_works_count: :desc)
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
