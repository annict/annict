# frozen_string_literal: true

module Cards
  class StatusCardComponent < ApplicationComponent
    def initialize(view_context, status:, page_category: "")
      super view_context
      @status = status
      @page_category = page_category
      @anime = @status.anime.decorate
    end

    def render
      build_html do |h|
        h.tag :div, class: "card" do
          h.tag :div, class: "card-body" do
            h.tag :div, class: "row g-3" do
              h.tag :div, class: "col-auto" do
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
                h.tag :div do
                  h.tag :span, class: "badge bg-#{@status.kind_v3}" do
                    h.text t("enumerize.status.kind_v3.#{@status.kind_v3}")
                  end
                end

                h.tag :div, class: "mt-1" do
                  h.tag :a, href: view_context.anime_path(@anime.id), class: "fw-bold text-body" do
                    h.text @anime.local_title
                  end
                end

                h.tag :div, class: "mt-1" do
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
end
