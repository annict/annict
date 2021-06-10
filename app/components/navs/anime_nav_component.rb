# frozen_string_literal: true

module Navs
  class AnimeNavComponent < ApplicationV6Component
    def initialize(view_context, anime:)
      super view_context
      @anime = anime
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-subnav c-subnav--transparent" do
          h.tag :a, href: view_context.anime_path(anime_id: @anime.id), class: "c-subnav__link #{"c-subnav__link--active" if page_category.in?(%w[anime])}" do
            h.tag :div, class: "c-subnav__item" do
              h.text t("noun.detail")
            end
          end

          unless @anime.no_episodes?
            h.tag :a, href: view_context.episode_list_path(anime_id: @anime.id), class: "c-subnav__link #{"c-subnav__link--active" if page_category.in?(%w[episode episode-list])}" do
              h.tag :div, class: "c-subnav__item" do
                h.text t("noun.episodes")
              end
            end
          end

          h.tag :a, href: view_context.anime_record_list_path(anime_id: @anime.id), class: "c-subnav__link #{"c-subnav__link--active" if page_category.in?(%w[anime-record-list])}" do
            h.tag :div, class: "c-subnav__item" do
              h.text t("noun.records")
            end
          end
        end
      end
    end
  end
end
