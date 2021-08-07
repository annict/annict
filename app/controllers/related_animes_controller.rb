# frozen_string_literal: true

class RelatedAnimesController < ApplicationV6Controller
  include AnimeHeaderLoadable

  def index
    set_page_category PageCategory::RELATED_ANIME_LIST

    set_anime_header_resources

    @series_list = @anime
      .series_list
      .eager_load(series_animes: {anime: :anime_image})
      .only_kept
      .order("works.season_year ASC")
      .order("works.season_name ASC")
      .order(:created_at)
  end
end