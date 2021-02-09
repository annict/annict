# frozen_string_literal: true

module Canary
  module Resolvers
    class SlotsOnProgram < Canary::Resolvers::Base
      def resolve(viewer_untracked: nil, order_by: nil)
        program = object
        viewer = context[:viewer]
        order = Canary::OrderProperty.build(order_by)

        slots = program.slots.only_kept

        if viewer_untracked
          library_entry = viewer.library_entries.find_by(work: program.work)

          if library_entry
            slots = slots.where.not(episode_id: library_entry.watched_episode_ids)
          end
        end

        slots.order(order.field => order.direction)
      end
    end
  end
end
