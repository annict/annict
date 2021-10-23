# frozen_string_literal: true

module Beta
  module Resolvers
    class LibraryEntries < Beta::Resolvers::Base
      def resolve(states: nil, until_current_season: nil, order_by: nil)
        order = Beta::OrderProperty.build(order_by)
        user = object

        library_entries = user.library_entries.with_not_deleted_work

        if states
          library_entries = library_entries.with_status(states.map(&:downcase))
        end

        if until_current_season
          library_entries = library_entries.joins(:work).merge(Work.lt_current_season)
        end

        case order.field
        when :last_tracked_at
          library_entries.order(position: order.direction == :asc ? :desc : :asc)
        else
          library_entries.order(created_at: :asc)
        end
      end
    end
  end
end
