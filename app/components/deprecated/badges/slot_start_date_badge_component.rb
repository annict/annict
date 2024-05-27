# typed: false
# frozen_string_literal: true

module Deprecated::Badges
  class SlotStartDateBadgeComponent < Deprecated::ApplicationV6Component
    def initialize(view_context, started_at:, time_zone:, class_name: "")
      super view_context
      @started_at = started_at
      @time_zone = time_zone
      @class_name = class_name
      @tv_time = TvTime.new(time_zone: @time_zone)
    end

    def render
      return "" if @started_at.nil?

      build_html do |h|
        h.tag :div, class: badge_class_name do
          h.text badge_text
        end
      end
    end

    private

    def badge_class_name
      class_name = %w[badge]
      class_name += @class_name.split(" ")
      class_name << "u-bg-slot-#{badge_name}" if badge_name
      class_name.join(" ")
    end

    def badge_text
      case badge_name
      when :yesterday then t("noun.yesterday")
      when :today then t("noun.today")
      when :tomorrow then t("noun.tomorrow")
      when :finished then t("noun.finished")
      end
    end

    def badge_name
      @badge_name ||= @tv_time.status_on(@started_at)
    end
  end
end
