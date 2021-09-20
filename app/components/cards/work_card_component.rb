# frozen_string_literal: true

module Cards
  class WorkCardComponent < ApplicationV6Component
    def initialize(view_context, work:, width:, caption: "", show_button: true, user: nil)
      super view_context
      @work = work
      @width = width
      @caption = caption
      @show_button = show_button
      @user = user
    end

    def render
      build_html do |h|
        h.tag :div, class: "border-0 card h-100" do
          h.tag :a, href: view_context.work_path(@work), class: "text-reset" do
            h.html Pictures::WorkPictureComponent.new(
              view_context,
              work: @work,
              width: @width,
              alt: @work.local_title
            ).render

            h.tag :div, class: "fw-bold h5 mb-0 mt-2 text-center text-truncate", title: @work.local_title do
              h.text @work.local_title
            end

            if @caption.present?
              h.tag :div, class: "mb-0 small text-center text-muted text-truncate", title: @caption do
                h.text @caption
              end
            end
          end

          if @show_button && (!current_user || !@user || current_user.id == @user.id)
            h.tag :div, class: "mt-2 text-center" do
              h.html ButtonGroups::WorkButtonGroupComponent.new(view_context, work: @work).render
            end
          end
        end
      end
    end
  end
end
