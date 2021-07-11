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
            h.html Pictures::AnimePictureComponent.new(
              view_context,
              anime: @anime,
              width: @width,
              alt: @anime.local_title
            ).render

            h.tag :div, class: "fw-bold h5 mb-0 mt-2 text-center text-truncate" do
              h.text @anime.local_title
            end
          end

          h.tag :div, class: "mt-2 text-center" do
            h.html ButtonGroups::AnimeButtonGroupComponent.new(view_context, anime: @anime).render
          end
        end
      end
    end
  end
end
