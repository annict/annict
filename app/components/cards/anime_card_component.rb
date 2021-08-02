# frozen_string_literal: true

module Cards
  class AnimeCardComponent < ApplicationV6Component
    def initialize(view_context, anime:, width:, caption: "", show_button: true, user: nil)
      super view_context
      @anime = anime
      @width = width
      @caption = caption
      @show_button = show_button
      @user = user
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

            if @caption.present?
              h.tag :div, class: "mb-0 small text-center text-muted text-truncate" do
                h.text @caption
              end
            end
          end

          if @show_button && (!current_user || !@user || current_user.id == @user.id)
            h.tag :div, class: "mt-2 text-center" do
              h.html ButtonGroups::AnimeButtonGroupComponent.new(view_context, anime: @anime).render
            end
          end
        end
      end
    end
  end
end
