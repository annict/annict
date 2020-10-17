# frozen_string_literal: true

module Canary
  module Resolvers
    class Slots < Canary::Resolvers::Base
      def resolve(until_next_night: nil, viewer_unwatched_only: nil, order_by: nil)
        user = object
        order = Canary::OrderProperty.build(order_by)

        works = user.works_on(:wanna_watch, :watching).only_kept
        user_programs = user.user_programs.where(work: works)
        slots = Slot.where(program_id: user_programs.pluck(:program_id))

        if until_next_night
          next_night = (Time.zone.now.in_time_zone(user.time_zone).end_of_day + 5.hours).utc
          slots = slots.before(next_night, field: :started_at)
        end

        if viewer_unwatched_only
          watched_episode_ids = user.library_entries.pluck(:watched_episode_ids).flatten
          slots = slots.where.not(episode_id: watched_episode_ids)
        end

        slots = slots.order(order.field => order.direction)

        slots
      end
    end
  end
end
