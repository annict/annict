# frozen_string_literal: true

module Headers
  class RecordHeaderComponent < ApplicationV6Component
    def initialize(view_context, record:, show_box: true, show_options: true)
      super view_context
      @record = record
      @show_options = show_options
      @show_box = show_box
      @user = @record.user
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-record-header row" do
          h.tag :div, class: "col-auto pe-0" do
            h.tag :a, href: view_context.profile_path(@user.username), target: "_top" do
              h.html view_context.render(Pictures::AvatarPictureComponent.new(user: @user, width: 50))
            end
          end

          h.tag :div, class: "col" do
            h.tag :div do
              h.tag :a, href: view_context.profile_path(@user.username), class: "fw-bold me-1 text-body", target: "_top" do
                h.tag :span, class: "me-1" do
                  h.text @user.name
                end

                h.tag :small, class: "text-muted" do
                  h.text "@#{@user.username}"
                end
              end

              if @user.supporter? && !@user.hide_supporter_badge?
                h.html Badges::SupporterBadgeComponent.new(view_context, user: @user).render
              end
            end

            h.tag :div do
              h.tag :a, href: view_context.record_path(@user.username, @record.id), class: "small text-muted", target: "_top" do
                h.text display_time(@record.watched_at)
              end

              if @record.modified_at
                h.tag :small, class: "ms-1 text-muted" do
                  h.tag :i, class: "fas fa-pencil-alt"
                end
              end
            end
          end

          h.tag :div, class: "col-auto ps-0" do
            if @show_options && current_user && RecordPolicy.new(current_user, @record).update?
              h.html Dropdowns::RecordOptionsDropdownComponent.new(view_context, record: @record, show_box: @show_box, show_options: @show_options).render
            end
          end
        end
      end
    end
  end
end
