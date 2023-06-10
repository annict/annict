# frozen_string_literal: true

module Fragment
  class TrackingHeatmapsController < Fragment::ApplicationController
    def show
      user = User.only_kept.find_by!(username: params[:username])
      date_from = (Date.today - 150.days).beginning_of_week(:sunday)
      count_data = user
        .records
        .after(date_from)
        .group_by_day(:watched_at, time_zone: time_zone)
        .count
        .map { |date, count| [date.to_s(:ymd), count] }
        .to_h
      @tracking_data = (date_from..(Date.today)).each_with_object({}) do |date, hash|
        formatted_date = date.to_s(:ymd)
        count = count_data[formatted_date]
        hash[formatted_date] = {
          count: count,
          leveled_count: leveled_count(count)
        }
      end
    end

    private

    def time_zone
      @time_zone ||= current_user&.time_zone.presence || cookies["ann_time_zone"].presence || "Asia/Tokyo"
    end

    def leveled_count(count)
      case count
      when 1..3 then 1
      when 4..6 then 2
      when 7..9 then 3
      when 10.. then 4
      else
        0
      end
    end
  end
end
