# frozen_string_literal: true

class AnimeSubNavComponent < ApplicationComponent
  def initialize(anime:, page_category:)
    @anime = anime
    @page_category = page_category.to_s
  end
end
