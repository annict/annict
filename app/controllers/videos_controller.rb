# frozen_string_literal: true

class VideosController < ApplicationV6Controller
  include AnimeHeaderLoadable

  def index
    set_page_category PageCategory::VIDEO_LIST

    set_anime_header_resources

    @videos = @anime.trailers.only_kept.order(:sort_number)
  end
end
