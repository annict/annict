# frozen_string_literal: true

module Cards
  class EpisodeRecordCardComponent < ApplicationComponent
    def initialize(view_context, episode_record:)
      super view_context
      @episode_record = episode_record
      @anime = @episode_record.anime.decorate
      @episode = @episode_record.episode
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-episode-record-card c-record-card card col-11 col-lg-10 mb-2 px-0" do
          h.tag :div, class: "card-body p-2" do
            h.tag :div, class: "row" do
              h.tag :div, class: "col-auto pr-0" do
                h.tag :a, href: view_context.anime_path(@anime.id) do
                  h.html Pictures::AnimePictureComponent.new(
                    view_context,
                    anime: @anime,
                    width: 80,
                    mb_width: 60
                  ).render
                end
              end

              h.tag :div, class: "col" do
                h.tag :div, class: "mb-1" do
                  h.tag :a, href: view_context.anime_path(@anime.id), class: "fw-bold text-body" do
                    h.text @anime.local_title
                  end

                  h.tag :a, href: view_context.episode_path(@anime.id, @episode.id), class: "fw-bold small text-body" do
                    h.tag :span, class: "pr-1" do
                      h.text @episode.local_number
                    end

                    h.text @episode.local_title
                  end
                end

                h.html Selectors::StatusSelectorComponent.new(
                  view_context,
                  anime: @anime,
                  page_category: @page_category
                ).render
              end
            end
          end
        end
      end
    end
  end
end
