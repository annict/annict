# frozen_string_literal: true

class WelcomeController < ApplicationController
  layout "welcome"

  def show
    set_page_category PageCategory::WELCOME

    anime_list = Anime.only_kept.preload(:work_image)
    @seasonal_anime_list = anime_list.by_season(ENV.fetch("ANNICT_CURRENT_SEASON")).order(watchers_count: :desc).limit(6)
    @popular_anime_list = anime_list.order(watchers_count: :desc).limit(6)

    @cover_image_anime = @seasonal_anime_list.sample
  end
end
