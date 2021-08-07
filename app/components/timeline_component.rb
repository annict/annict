# frozen_string_literal: true

class TimelineComponent < ApplicationV6Component
  def initialize(view_context, activity_groups:)
    super view_context
    @activity_groups = activity_groups
  end

  def render
    build_html do |h|
      h.tag :div, class: "c-timeline" do
        h.tag :div, class: "c-timeline__activities" do
          @activity_groups.each.with_prelude do |activity_group|
            h.tag :div, class: "c-timeline__activity pt-3" do
              case activity_group.itemable_type
              when "Status"
                h.html Activities::StatusActivityComponent.new(view_context, activity_group: activity_group).render
              when "EpisodeRecord", "WorkRecord", "AnimeRecord"
                h.html Activities::RecordActivityComponent.new(view_context, activity_group: activity_group).render
              end
            end
          end
        end

        h.tag :div, class: "mt-3 text-center" do
          h.html ButtonGroups::PaginationButtonGroupComponent.new(view_context, collection: @activity_groups).render
        end
      end
    end
  end
end
