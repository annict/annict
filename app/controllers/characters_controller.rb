# frozen_string_literal: true

class CharactersController < ApplicationController
  before_action :load_i18n, only: %i(show)

  def show
    @character = Character.published.find(params[:id])
    @casts_with_year = @character.
      casts.
      published.
      joins(:work).
      where(works: { aasm_state: :published }).
      group_by { |cast| cast.work.season_year.presence || 0 }
    @cast_years = @casts_with_year.keys.sort.reverse

    @favorite_characters = @character.
      favorite_characters.
      joins(:user).
      merge(User.published).
      order(id: :desc)
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
