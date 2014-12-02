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
        sql = 'date(created_at) as checkins_day, count(*) as checkins_count'
        weekly_checkins = user.checkins.where(created_at: week_days)
        weekly_checkins = weekly_checkins
                            .select(sql)
                            .group('date(created_at)')

        weekly_checkins_hash = {}
        weekly_checkins.each do |checkin|
          checkin_date = checkin.checkins_day.strftime('%-m/%d')
          weekly_checkins_hash[checkin_date] = checkin.checkins_count
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
