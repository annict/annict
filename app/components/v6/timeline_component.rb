# frozen_string_literal: true

module V6
  class TimelineComponent < V6::ApplicationComponent
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
              h.tag :div, class: "c-timeline__activity container u-container-flat pt-3" do
                case activity_group.itemable_type
                when "Status"
                  h.html V6::Activities::StatusActivityComponent.new(
                    view_context,
                    activity_group: activity_group,
                    page_category: @page_category
                  ).render
                when "EpisodeRecord", "WorkRecord"
                  h.html V6::Activities::RecordActivityComponent.new(
                    view_context,
                    activity_group: activity_group,
                    page_category: @page_category
                  ).render
                end
              end
            end
          end

          h.tag :div, class: "mt-3 text-center" do
            h.html V6::ButtonGroups::PaginationButtonGroupComponent.new(view_context, collection: @activity_groups).render
          end
        end
      end
    end
  end
end
