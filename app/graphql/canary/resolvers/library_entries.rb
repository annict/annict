# frozen_string_literal: true

module Canary
  module Resolvers
    class LibraryEntries < Canary::Resolvers::Base
      def resolve(status_kinds: nil, until_current_season: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)
        user = object

        library_entries = user.library_entries.with_not_deleted_work

        if status_kinds
          library_entries = library_entries.with_status(status_kinds.map { |kind| Status.kind_v3_to_v2(kind.downcase) })
        end

        if until_current_season
          library_entries = library_entries.joins(:work).merge(Work.until_current_season)
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
