# frozen_string_literal: true

class FavoriteOrganizationsController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index(username)
    @user = User.find_by!(username: username)
    @favorite_organizations = @user.
      favorite_organizations.
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
