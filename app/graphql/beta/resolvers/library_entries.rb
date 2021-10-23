# frozen_string_literal: true

module Beta
  module Resolvers
    class LibraryEntries < Beta::Resolvers::Base
      def resolve(states: nil, seasons: nil, season_from: nil, season_until: nil, order_by: nil)
        order = Beta::OrderProperty.build(order_by)
        user = object

        library_entries = user.library_entries.with_not_deleted_work

        if states
          library_entries = library_entries.with_status(states.map(&:downcase))
        end

        if seasons
          season_slugs = find_seasons(seasons).map(&:slug)
          library_entries = library_entries.joins(:work).merge(Work.by_seasons(season_slugs))
        end

        if season_from
          library_entries = library_entries.joins(:work).merge(Work.season_from(find_season(season_from)))
        end

        if season_until
          library_entries = library_entries.joins(:work).merge(Work.season_until(find_season(season_until)))
        end

        case order.field
        when :last_tracked_at
          library_entries.order(position: order.direction == :asc ? :desc : :asc)
        else
          library_entries.order(created_at: :asc)
        end
      end

      private

      def find_season(season_slug)
        Season.find_by_slug(season_slug).presence || raise(GraphQL::ExecutionError, "Invalid season: #{season_slug}")
      end

      def find_seasons(season_slugs)
        seasons = season_slugs.map { |slug| Season.find_by_slug(slug) }

        if seasons.any?(&:nil?)
          raise GraphQL::ExecutionError, "Invalid seasons: #{season_slugs}"
        end

        seasons
      end
    end
  end
end
