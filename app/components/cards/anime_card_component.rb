# frozen_string_literal: true

module Cards
  class AnimeCardComponent < ApplicationComponent
    def initialize(view_context, anime:, width:, mb_width:, page_category: "")
      super view_context
      @anime = anime
      @width = width
      @mb_width = mb_width
      @page_category = page_category
    end

    def render
      build_html do |h|
        h.tag :div, class: "border-0 card h-100" do
          h.tag :a, href: view_context.anime_path(@anime), class: "text-reset" do
            h.html Pictures::AnimePictureComponent.new(
              view_context,
              anime: @anime,
              width: @width,
              mb_width: @mb_width,
              alt: @anime.local_title
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
