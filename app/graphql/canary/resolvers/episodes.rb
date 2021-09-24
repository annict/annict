# frozen_string_literal: true

module Canary
  module Resolvers
    class Episodes < Canary::Resolvers::Base
      include GraphQL::FragmentCache::ObjectHelpers

      def resolve(viewer_tracked_in_current_status: nil, order_by: nil)
        cache_fragment(query_cache_key: query_cache_key(context, object)) do
          order = Canary::OrderProperty.build(order_by)
          viewer = context[:viewer]
          work = object
          episodes = work.episodes.only_kept

          if viewer && viewer_tracked_in_current_status
            library_entry = viewer.library_entries.with_not_deleted_work.find_by(work: work)

            if library_entry
              episodes = episodes.where(id: library_entry.watched_episode_ids)
            end
          elsif viewer && viewer_tracked_in_current_status == false
            library_entry = viewer.library_entries.with_not_deleted_work.find_by(work: work)

            if library_entry
              episodes = episodes.where(id: episodes.pluck(:id) - library_entry.watched_episode_ids)
            end
          end

          case order.field
          when :created_at, :sort_number
            episodes.order(order.field => order.direction)
          else
            episodes
          end
        end
      end

      private

      def query_cache_key(context, object)
        [
          Digest::SHA1.hexdigest(context.query.query_string),
          self.class.name,
          context[:viewer]&.id.inspect,
          context[:viewer]&.status_cache_expired_at.to_i,
          context[:viewer]&.record_cache_expired_at.to_i,
          object.updated_at.to_i,
          "1" # キャッシュクリア用の任意の文字列
        ].freeze
      end
    end
  end
end
