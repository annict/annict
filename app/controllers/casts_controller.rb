# frozen_string_literal: true

class CastsController < ApplicationV6Controller
  include AnimeHeaderLoadable

  def index
    set_page_category PageCategory::CAST_LIST

    set_anime_header_resources

    @casts = @anime.casts.only_kept.order(:sort_number)
  end
end
