# frozen_string_literal: true

module V6::Boxes
  class AnimeBoxComponent < V6::ApplicationComponent
    def initialize(view_context, anime:, page_category:, episode: nil)
      super view_context
      @anime = anime
      @page_category = page_category
      @episode = episode
    end

    def render
      build_html do |h|
        h.tag :div, class: "row g-3" do
          h.tag :div, class: "col-auto" do
            h.tag :a, href: view_context.anime_path(@anime.id) do
              h.html V6::Pictures::AnimePictureComponent.new(
                view_context,
                anime: @anime,
                width: 80,
                mb_width: 40
              ).render
            end
          end

          h.tag :div, class: "col" do
            h.tag :div do
              h.tag :a, href: view_context.anime_path(@anime.id), class: "text-body" do
                h.text @anime.local_title
              end
            end

            if @episode
              h.tag :div do
                h.tag :a, href: view_context.episode_path(@anime.id, @episode.id), class: "fw-bold small text-body" do
                  h.tag :span, class: "px-1" do
                    h.text @episode.local_number
                  end

                  h.text @episode.local_title
                end
              end
            end

            h.html V6::Selectors::StatusSelectorComponent.new(
              view_context,
              anime: @anime,
              page_category: @page_category,
              class_name: "mt-1"
            ).render
          end
        end
      end
    end
  end
end