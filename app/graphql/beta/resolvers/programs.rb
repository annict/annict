# frozen_string_literal: true

module Beta
  module Resolvers
    class Programs < Beta::Resolvers::Base
      def resolve(unwatched: nil, order_by: nil)
        @user = object
        @watched = unwatched.nil? ? nil : !unwatched

        order = Beta::OrderProperty.build(order_by)

        collection = user_slots.preload(:channel, work: :work_image, episode: :work)
        collection = collection.order(order.field => order.direction)

        collection
      end

      private

      attr_reader :user, :watched

      def slots
        Slot.only_kept.with_works(user.works_on(:wanna_watch, :watching).only_kept)
      end

      def user_slots
        return slots.none if user_episodes.blank?

        id_pair = channel_works.pluck(:channel_id, :work_id).map { |ary| "(#{ary[0]}, #{ary[1]})" }.join(",")
        id_pair_sql = id_pair.blank? ? "NULL" : <<-SQL
      SELECT id FROM slots WHERE
        (channel_id, work_id) IN (VALUES #{id_pair})
        SQL

        # 過去の放送を取得しないようにするため、直近の放送分のみ取得する
        slot_sql = <<-SQL
      WITH ranked_slots AS (
        SELECT id, episode_id,
          dense_rank() OVER (
            PARTITION BY episode_id ORDER BY started_at DESC
          ) AS episode_rank
        FROM slots
        WHERE
          id IN (#{id_pair_sql})
      )
      SELECT id FROM ranked_slots
      WHERE
        episode_id IN (#{user_episodes.pluck(:id).join(',')}) AND
        episode_rank = 1;
        SQL
        slots_ids = slots.find_by_sql(slot_sql)

        slots.where(id: slots_ids.pluck(:id))
      end

      def user_episodes
        @user_episodes ||= UserEpisodesQuery.new(
          user,
          Episode.only_kept.where(work_id: library_entries.pluck(:work_id)),
          watched: watched
        ).call
      end

      def channel_works
        @channel_works ||= user.channel_works.where(work: library_entries.pluck(:work_id))
      end

      def library_entries
        @library_entries ||= user.library_entries
      end

      def order_collection(collection)
        return collection.order(:created_at) unless order

        collection.order(order.field => order.direction)
      end
    end
  end
end
