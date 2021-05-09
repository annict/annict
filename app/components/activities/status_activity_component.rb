# frozen_string_literal: true

module Activities
  class StatusActivityComponent < ApplicationComponent
    def initialize(view_context, activity_struct:, page_category: "")
      super view_context
      @activity_struct = activity_struct
      @page_category = page_category
      @user = activity_group_struct.user.decorate
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-timeline__status-activity" do
          h.tag :div, class: "row g-3" do
            h.tag :div, class: "col-auto" do
              h.tag :a, href: view_context.profile_path(@user.username) do
                h.html Pictures::AvatarPictureComponent.new(view_context,
                  user: @user,
                  width: 32,
                  mb_width: 32
                ).render
              end
            end

            h.tag :div, class: "col" do
              h.tag :div do
                h.tag :span, class: "c-timeline__user-name" do
                  h.tag :a, href: view_context.profile_path(@user.username), class: "text-body fw-bold u-link" do
                    h.text @user.name
                  end
                end

                h.tag :small, class: "ms-1" do
                  h.text t("messages._components.activities.status.changed")
                end

                h.html RelativeTimeComponent.new(
                  view_context,
                  time: @activity_group_struct.created_at.iso8601,
                  class_name: "ms-1 small text-muted"
                ).render
              end

              h.tag :div, class: "c-timeline__activity-cards" do
                @activity_group_struct.itemables.each do |status|
                  h.tag :div, class: "mt-3" do
                    h.html Contents::StatusContentComponent.new(
                      view_context,
                      status: status,
                      page_category: @page_category
                    ).render
                  end
                end
              end

              if @activity_group_struct.itemables.length > 2
                h.tag :div, class: "text-center" do
                  h.tag :a, class: "c-activity-more-button btn btn-outline-secondary btn-small py-1", href: "" do
                    h.tag :i, class: "fal fa-chevron-double-down me-1"
                    h.text t("messages._components.activities.status.more", n: @activity_group_struct.itemables.length - 2)
                  end
                end
              end
            end
          end
        end
      end
    end
  end
end
