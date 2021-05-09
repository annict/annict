# frozen_string_literal: true

class TimelineComponent < ApplicationComponent
  def initialize(view_context, activity_groups:, page_category:)
    super view_context
    @activity_groups = activity_groups
    @page_category = page_category
  end

  def render
    build_html do |h|
      h.tag :div, class: "c-timeline" do
        h.tag :div, class: "c-timeline__activities" do
          @activity_groups.each.with_prelude do |activity_group|
            h.tag :div, class: "c-timeline__activity py-3 u-underline" do
              case activity_group.itemable_type
              when "Status"
                h.html Activities::StatusActivityComponent.new(
                  view_context,
                  activity_group: activity_group,
                  page_category: @page_category
                ).render
              # when "record"
              #   h.html Activities::RecordActivityComponent.new(
              #     view_context,
              #     activity_group_struct: activity_group_struct,
              #     page_category: @page_category
              #   ).render
              end
            end
          end
        end

        if @activity_groups.total_pages > 1
          h.tag :div, class: "mt-3 text-center" do
            h.html paginate(@activity_groups)
          end
        end
      end
    end
  end
end
