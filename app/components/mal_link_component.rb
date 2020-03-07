# frozen_string_literal: true

class MALLinkComponent < ApplicationComponent
  inline!

  def initialize(work:, title: nil)
    @work = work
    @title = title
  end

  def call
    return "-" if work.mal_anime_id.blank?

    link_to link_title, work.mal_anime_url, target: "_blank", rel: "noopener"
  end

  private

  attr_reader :title, :work

  def link_title
    title.presence || work.mal_anime_id
  end
end
