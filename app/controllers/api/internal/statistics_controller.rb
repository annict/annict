# frozen_string_literal: true

module Api
  module Internal
    class StatisticsController < Api::Internal::ApplicationController
      def user_heatmap(username, start_date, end_date)
        start_date = Time.parse(start_date)
        end_date = Time.parse(end_date)
        user = User.find_by!(username: username)
        @days = user.activities.records_and_reviews.between_times(start_date, end_date).
          group_by_day(:created_at).count.
          map { |date, val| [date.to_time.to_i, val] }.
          to_h
      end
    end
  end
end
