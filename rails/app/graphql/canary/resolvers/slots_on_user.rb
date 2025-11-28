# typed: false
# frozen_string_literal: true

module Canary
  module Resolvers
    class SlotsOnUser < Canary::Resolvers::Base
      def resolve(until_next_night: nil, untracked: nil, order_by: nil)
        user = object
        order = Canary::OrderProperty.build(order_by)

        library_entries = user.library_entries.wanna_watch_and_watching
        slots = Slot.where(program: library_entries.select(:program_id))

        if until_next_night
          tv_time = TvTime.new(time_zone: user.time_zone)
          next_night = (tv_time.beginning_of_today.tomorrow.end_of_day + 5.hours).utc
          slots = slots.before(next_night, field: :started_at)
        end

        if untracked
          watched_episode_ids = library_entries.pluck(:watched_episode_ids).flatten
          slots = slots.where.not(episode_id: watched_episode_ids)
        end

        slots.order(order.field => order.direction)
      end
    end
  end
end
