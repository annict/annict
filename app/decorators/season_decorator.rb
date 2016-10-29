# frozen_string_literal: true

class SeasonDecorator < ApplicationDecorator
  def local_name
    h.season_local_name(year, name)
  end
end
