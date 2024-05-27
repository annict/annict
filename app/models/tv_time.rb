# typed: false
# frozen_string_literal: true

class TvTime
  def initialize(time_zone:)
    @time_zone = time_zone
  end

  def now
    Time.zone.now.in_time_zone(@time_zone)
  end

  def beginning_of_today
    now.hour <= 4 ? now.yesterday.beginning_of_day : now.beginning_of_day
  end

  def status_on(time)
    if (beginning_of_today - 1.day + 5.hours) < time && time < (beginning_of_today.end_of_day - 1.day + 5.hours)
      :yesterday
    elsif (beginning_of_today + 5.hours) < time && time < (beginning_of_today.end_of_day + 5.hours)
      :today
    elsif (beginning_of_today + 1.day + 5.hours) < time && time < (beginning_of_today.end_of_day + 1.day + 5.hours)
      :tomorrow
    elsif time < (beginning_of_today - 1.day + 5.hours)
      :finished
    end
  end
end
