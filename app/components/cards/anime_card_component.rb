# frozen_string_literal: true

module Cards
  class AnimeCardComponent < ApplicationV6Component
    def initialize(view_context, anime:, width:)
      super view_context
      @anime = anime
      @width = width
    end

    def render
      build_html do |h|
        h.tag :div, class: "border-0 card h-100" do
          h.tag :a, href: view_context.anime_path(@anime), class: "text-reset" do
            h.html V6::Pictures::AnimePictureComponent.new(
              view_context,
              anime: @anime,
              width: @width,
              alt: @anime.local_title
            ).render

            h.tag :h5, class: "fw-bold mb-0 mt-2 text-truncate" do
              h.text @anime.local_title
            end
          end

          h.html V6::Selectors::StatusSelectorComponent.new(
            view_context,
            anime: @anime,
            class_name: "mt-2"
          ).render
        end
      end
    end
  end
end
