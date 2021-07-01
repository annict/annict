# frozen_string_literal: true

class MalLinkComponent < ApplicationComponent
  def initialize(anime:, title: nil)
    @anime = anime
    @title = title
  end

  def call
    return "-" if @anime.mal_anime_id.blank?

    link_to link_title, @anime.mal_anime_url, target: "_blank", rel: "noopener"
  end

  private

  def link_title
    @title.presence || @anime.mal_anime_id
  end
end
