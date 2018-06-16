# frozen_string_literal: true

module Api
  module Internal
    class StatisticsController < Api::Internal::ApplicationController
      include TimeZoneHelper

      def user_heatmap(username, start_date, end_date)
        start_date = Time.parse(start_date)
        end_date = Time.parse(end_date)
        user = User.find_by!(username: username)
        time_zone = local_time_zone.presence || user.time_zone

        @days = user.records.between_times(start_date, end_date).
          group_by_day(:created_at, time_zone: time_zone).count.
          map { |date, val| [date.to_time.to_i, val] }.
          to_h
      end
    end
  end
end
