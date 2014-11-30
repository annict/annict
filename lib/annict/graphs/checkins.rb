module Annict
  module Graphs
    class Checkins
      def self.labels
        today = Date.today
        week_days = (today - 6.days)..today

        week_days.map { |day| day.strftime('%-m/%d') }
      end

      def self.values(user)
        today = Date.today
        week_days = (today - 6.days).beginning_of_day..today.end_of_day
        weekly_checkins = user.checkins.where(created_at: week_days)
        weekly_checkins = weekly_checkins
                            .select('date(created_at) as checkins_day, count(*) as checkins_count')
                            .group('date(created_at)')

        weekly_checkins_hash = {}
        weekly_checkins.each do |checkin|
          weekly_checkins_hash[checkin.checkins_day.strftime('%-m/%d')] = checkin.checkins_count
        end

        checkins = []
        labels.each do |label|
          checkins << (weekly_checkins_hash[label].presence || 0)
        end

        checkins
      end
    end
  end
end
