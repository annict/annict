# frozen_string_literal: true

module Cards
  class StatusCardComponent < ApplicationComponent2
    def initialize(view_context, status:, page_category: "")
      super view_context
      @status = status
      @page_category = page_category
      @anime = @status.anime.decorate
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-status-card card col-11 col-lg-10 mb-2 px-0" do
          h.tag :div, class: "card-body p-2" do
            h.tag :div, class: "row" do
              h.tag :div, class: "col-auto pr-0" do
                h.tag :a, href: view_context.anime_path(@anime.id) do
                  h.html Images::AnimeImageComponent.new(
                    view_context,
                    image_url_1x: @anime.image_url(size: "350x"),
                    image_url_2x: @anime.image_url(size: "700x"),
                    alt: @anime.local_title
                  ).render
                end
              end

              h.tag :div, class: "col" do
                h.tag :div do
                  h.tag :span, class: "badge u-badge-#{@status.kind}" do
                    h.text t("enumerize.status.kind_v3.#{@status.kind}")
                  end
                end

                h.tag :div, class: "mb-1" do
                  h.tag :a, href: view_context.anime_path(@anime.id) do
                    h.text @anime.local_title
                  end
                end

                h.html Selectors::StatusSelectorComponent.new(
                  view_context,
                  anime: @anime,
                  page_category: @page_category,
                  small: true
                ).render
              end
            end
          end
        end
      end
    end
  end
end
