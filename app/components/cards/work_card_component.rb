# frozen_string_literal: true

module Cards
  class WorkCardComponent < ApplicationV6Component
    def initialize(view_context, work:, width:, show_button: true, show_note: false, user: nil, caption: "", note: "")
      super view_context
      @work = work
      @width = width
      @show_button = show_button
      @show_note = show_note
      @user = user
      @caption = caption
      @note = note
    end

    def render
      build_html do |h|
        h.tag :div, class: "align-items-end border-0 c-work-card card flex-column h-100" do
          h.tag :div, class: "text-center w-100" do
            h.tag :a, href: view_context.work_path(@work) do
              h.html view_context.render(
                Pictures::WorkPictureComponent.new(
                  work: @work,
                  width: @width,
                  alt: @work.local_title
                )
              )
            end
          end

          h.tag :div, class: "mt-2 w-100" do
            h.tag :a, href: view_context.work_path(@work), class: "text-body" do
              h.tag :div, class: "c-work-card__work-title fw-bold h5 mb-0", title: @work.local_title do
                h.text @work.local_title
              end

              if @caption.present?
                h.tag :div, class: "small text-muted text-truncate", title: @caption do
                  h.text @caption
                end
              end
            end
          end

          h.tag :div, class: "mt-auto text-center w-100" do
            if @note.present?
              h.tag :div, class: "mb-2" do
                h.tag :a, {
                  class: "px-2 small u-fake-link",
                  data_bs_toggle: "popover",
                  data_bs_trigger: "focus",
                  data_bs_placement: "top",
                  data_bs_content: @note,
                  tabindex: "0"
                } do
                  h.tag :i, class: "far fa-comment-dots me-1"
                  h.text t("noun.note_alt")
                end
              end
            end

            if @show_button && (!current_user || !@user || current_user.id == @user.id)
              h.tag :div, class: "mt-2" do
                h.html ButtonGroups::WorkButtonGroupComponent.new(view_context, work: @work).render
              end
            end
          end
        end
      end
    end
  end
end
