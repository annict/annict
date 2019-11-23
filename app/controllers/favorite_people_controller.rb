# frozen_string_literal: true

class FavoritePeopleController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index
    @user = User.without_deleted.find_by!(username: params[:username])
    @favorite_casts = @user.
      favorite_people.
      with_cast.
      order(watched_works_count: :desc)
    @favorite_staffs = @user.
      favorite_people.
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
