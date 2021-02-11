# frozen_string_literal: true

module Canary
  module Resolvers
    class Episodes < Canary::Resolvers::Base
      def resolve(viewer_checked_in_current_status: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)
        viewer = context[:viewer]
        anime = object

        episode_ids = Rails.cache.fetch(cache_key(viewer, anime, viewer_checked_in_current_status), expires_in: 24.hours) do
          episodes = anime.episodes.only_kept

          if viewer_checked_in_current_status
            library_entry = viewer.library_entries.with_not_deleted_work.find_by(work: anime)

            if library_entry
              episodes = episodes.where(id: library_entry.watched_episode_ids)
            end
          elsif viewer_checked_in_current_status == false
            library_entry = viewer.library_entries.with_not_deleted_work.find_by(work: anime)

            if library_entry
              episodes = episodes.where(id: episodes.pluck(:id) - library_entry.watched_episode_ids)
            end
          end

          episodes.pluck(:id)
        end

        episodes = Episode.where(id: episode_ids)

        case order.field
        when :created_at, :sort_number
          episodes.order(order.field => order.direction)
        else
          episodes
        end
      end

      private

      def cache_key(viewer, anime, viewer_checked_in_current_status)
        [
          self.class.name,
          viewer.id,
          anime.id,
          anime.updated_at.rfc3339,
          viewer_checked_in_current_status.inspect,
          '1' # キャッシュクリア用の任意の文字列
        ].freeze
      end
    end
  end
end
