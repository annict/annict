# frozen_string_literal: true

module Annict
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
  end
end
