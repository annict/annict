# frozen_string_literal: true

class AnimeInfosController < ApplicationV6Controller
  include AnimeHeaderLoadable

  def show
    set_page_category PageCategory::ANIME_INFO

    set_anime_header_resources
  end
end
