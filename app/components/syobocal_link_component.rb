# frozen_string_literal: true

class SyobocalLinkComponent < ApplicationComponent
  def initialize(anime:, title: nil)
    @anime = anime
    @title = title
  end

  def call
    return "-" if @anime.syobocal_tid.blank?

    link_to link_title, @anime.syobocal_url, target: "_blank", rel: "noopener"
  end

  private

  def link_title
    @title.presence || @anime.syobocal_tid
  end
end
