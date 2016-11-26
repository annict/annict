# frozen_string_literal: true

class CharactersController < ApplicationController
  before_action :load_work, only: %i(index)

  def index
    @casts = @work.casts.published.order(:sort_number)
  end
end
