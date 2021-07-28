# frozen_string_literal: true

module Boxes
  class AnimeBoxComponent < ApplicationV6Component
    def initialize(view_context, anime:, episode: nil)
      super view_context
      @anime = anime
      @episode = episode
    end

    def render
      build_html do |h|
        h.tag :div, class: "row g-3" do
          h.tag :div, class: "col-auto" do
            h.tag :a, href: view_context.anime_path(@anime.id), target: "_top" do
              h.html Pictures::AnimePictureComponent.new(
                view_context,
                anime: @anime,
                width: 80
              ).render
            end
          end

          h.tag :div, class: "col" do
            h.tag :div do
              h.tag :a, href: view_context.anime_path(@anime.id), class: "text-body", target: "_top" do
                h.text @anime.local_title
              end
            end

            if @episode
              h.tag :div do
                h.tag :a, href: view_context.episode_path(@anime.id, @episode.id), class: "fw-bold small text-body", target: "_top" do
                  h.tag :span, class: "px-1" do
                    h.text @episode.local_number
                  end

                  h.text @episode.local_title
                end
              end
            end

            h.tag :div, class: "mt-1" do
              h.html ButtonGroups::AnimeButtonGroupComponent.new(view_context, anime: @anime).render
            end
          end
        end
      end
    end
  end
end
