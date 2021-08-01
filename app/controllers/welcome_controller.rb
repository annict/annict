# frozen_string_literal: true

class WelcomeController < ApplicationV6Controller
  layout "main_welcome"

  def show
    set_page_category PageCategory::WELCOME

    anime_list = Anime.only_kept.preload(:anime_image)
    @seasonal_anime_list = anime_list.by_season(ENV.fetch("ANNICT_CURRENT_SEASON")).order(watchers_count: :desc).limit(6)
    cover_image_path = @seasonal_anime_list.sample&.anime_image&.uploaded_file_path(:image)
    @cover_image_url = cover_image_path ? helpers.ix_image_url(cover_image_path, blur: 60, height: 800, width: 600) : ""
  end
end
