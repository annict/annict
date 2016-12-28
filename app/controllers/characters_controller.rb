# frozen_string_literal: true

class CharactersController < ApplicationController
  before_action :load_work, only: %i(index)

  def index
    @casts = @work.casts.published.order(:sort_number)
  end

  def show(id)
    @character = Character.published.find(id)
    @casts_with_year = @character.
      casts.
      published.
      group_by { |cast| cast.work.season&.year.presence || 0 }
    @cast_years = @casts_with_year.keys.sort.reverse
  end
end
