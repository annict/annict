# frozen_string_literal: true

class PopularAnimesController < ApplicationV6Controller
  include AnimeListable

  def index
    set_page_category PageCategory::POPULAR_ANIME_LIST

    @animes = Anime
      .only_kept
      .preload(:anime_image)
      .order(watchers_count: :desc, id: :desc)
      .page(params[:page])
      .per(display_works_count)

    set_resource_data(@animes)
  end
end
