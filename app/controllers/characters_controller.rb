# frozen_string_literal: true

class CharactersController < ApplicationController
  before_action :load_i18n, only: %i(show)
  before_action :load_work, only: %i(index)

  def index
    @casts = @work.casts.published.order(:sort_number)

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end

  def show(id)
    @character = Character.published.find(id)
    @casts_with_year = @character.
      casts.
      published.
      group_by { |cast| cast.work.season&.year.presence || 0 }
    @cast_years = @casts_with_year.keys.sort.reverse

    @fan_users = @character.users.order("favorite_characters.id DESC")
  end

  private

  def load_i18n
    keys = {
      "messages.components.favorite_button.add_to_favorites": nil,
      "messages.components.favorite_button.added_to_favorites": nil,
    }

    load_i18n_into_gon keys
  end
end
