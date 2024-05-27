# typed: false
# frozen_string_literal: true

module Canary
  module Resolvers
    class SlotsOnProgram < Canary::Resolvers::Base
      include GraphQL::FragmentCache::ObjectHelpers

      def resolve(viewer_untracked: nil, order_by: nil)
        cache_fragment(query_cache_key: query_cache_key(context, object)) do
          viewer = context[:viewer]
          program = object
          order = Canary::OrderProperty.build(order_by)
          slots = program.slots.only_kept

          if viewer_untracked
            library_entry = viewer.library_entries.find_by(work_id: program.work_id)

            if library_entry
              slots = slots.where.not(episode_id: library_entry.watched_episode_ids)
            end
          end

          slots.order(order.field => order.direction)
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
