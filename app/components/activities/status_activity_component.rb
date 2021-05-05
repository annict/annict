# frozen_string_literal: true

module Activities
  class StatusActivityComponent < ApplicationComponent
    def initialize(view_context, activity_group_struct:, page_category: "")
      super view_context
      @activity_group_struct = activity_group_struct
      @page_category = page_category
      @user = activity_group_struct.user.decorate
    end

    def render
      build_html do |h|
        h.tag :div, class: "c-timeline__status-activity" do
          h.tag :div, class: "row mb-3" do
            h.tag :div, class: "col-auto pr-0" do
              h.tag :a, href: view_context.profile_path(@user.username) do
                h.html ProfileImageComponent2.new(view_context,
                  image_url_1x: @user.avatar_url(size: "50x50"),
                  alt: "@#{@user.username}",
                  lazy_load: false).render
              end
            end

            h.tag :div, class: "col" do
              h.tag :span, class: "c-timeline__user-name" do
                h.tag :a, href: view_context.profile_path(@user.username), class: "text-body font-weight-bold u-link" do
                  h.text @user.name
                end
              end

              h.tag :small, class: "mr-1" do
                h.text t("messages._components.activities.status.changed")
              end

              h.html RelativeTimeComponent.new(
                view_context,
                time: @activity_group_struct.created_at.iso8601,
                class_name: "small text-muted"
              ).render
            end
          end

          h.tag :div, class: "c-timeline__activity-cards" do
            @activity_group_struct.itemables.each do |status|
              h.tag :div, class: "mb-3" do
                h.html Contents::StatusContentComponent.new(
                  view_context,
                  status: status,
                  page_category: @page_category
                ).render
              end
            end
          end

          if @activity_group_struct.activities_count > 2
            h.tag :div, class: "text-center" do
              h.tag :a, class: "c-activity-more-button btn btn-outline-secondary btn-small py-1", href: "" do
                h.tag :i, class: "fal fa-chevron-double-down"
                h.text t("messages._components.activities.status.more", n: @activity_group_struct.activities_count - 2)
              end
            end
          end
        end
      end
    end
  end
end
