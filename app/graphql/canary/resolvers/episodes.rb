# frozen_string_literal: true

module Canary
  module Resolvers
    class Episodes < Canary::Resolvers::Base
      def resolve(viewer_checked_in_current_status: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)
        viewer = context[:viewer]

        @episodes = object.episodes.only_kept

        if viewer_checked_in_current_status
          library_entry = viewer.library_entries.with_not_deleted_work.find_by(work: object)

          if library_entry
            @episodes = @episodes.where(id: library_entry.watched_episode_ids)
          end
        elsif viewer_checked_in_current_status == false
          library_entry = viewer.library_entries.with_not_deleted_work.find_by(work: object)

          if library_entry
            @episodes = @episodes.where(id: @episodes.pluck(:id) - library_entry.watched_episode_ids)
          end
        end

        @episodes = case order.field
        when :created_at, :sort_number
          @episodes.order(order.field => order.direction)
        else
          @episodes
        end

        @episodes
      end
    end
  end
end
