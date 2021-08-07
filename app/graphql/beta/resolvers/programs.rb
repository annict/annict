# frozen_string_literal: true

module Beta
  module Resolvers
    class Programs < Beta::Resolvers::Base
      def resolve(unwatched: nil, order_by: nil)
        user = object
        order = Beta::OrderProperty.build(order_by)

        library_entries = user.library_entries.wanna_watch_and_watching
        slots = Slot.where(program: library_entries.select(:program_id))

        if unwatched
          watched_episode_ids = library_entries.pluck(:watched_episode_ids).flatten
          slots = slots.where.not(episode_id: watched_episode_ids)
        end

        slots.order(order.field => order.direction)
      end
    end
  end
end
