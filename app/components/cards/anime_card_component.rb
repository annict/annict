# frozen_string_literal: true

module Cards
  class AnimeCardComponent < ApplicationComponent2
    def initialize(view_context, anime:, page_category: "")
      super view_context
      @anime = anime
      @page_category = page_category
    end

    def render
      build_html do |h|
        h.tag :div, class: "border-0 card h-100" do
          h.tag :a, href: view_context.anime_path(@anime), class: "text-reset" do
            h.html Images::AnimeImageComponent.new(
              view_context,
              image_url_1x: @anime.image_url(size: "150x"),
              image_url_2x: @anime.image_url(size: "300x"),
              alt: @anime.local_title,
              class_name: "border"
            ).render

            h.tag :h5, class: "font-weight-bold mb-0 mt-2 text-truncate" do
              h.text @anime.local_title
            end
          end

          h.html Selectors::StatusSelectorComponent.new(
            view_context,
            anime: @anime,
            page_category: @page_category,
            class_name: "mt-2"
          ).render
        end
      end
    end
  end
end
