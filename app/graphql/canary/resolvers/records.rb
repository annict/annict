# frozen_string_literal: true

module Canary
  module Resolvers
    class Records < Canary::Resolvers::Base
      def resolve(has_body: nil, month: nil, episode_id: nil, by_following: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)

        @records = object.records.only_kept

        if has_body
          @records = @records.merge(EpisodeRecord.with_body)
        end

        if month
          unless %r{[0-9]{4}-[0-9]{2}}.match?(month)
            raise GraphQL::ExecutionError, "The `month` argument should be like `2020-03`"
          end

          start_time = Time.zone.parse("#{month}-01").in_time_zone(object.time_zone).beginning_of_month
          end_time = start_time.end_of_month
          @records = @records.between_times(start_time, end_time)
        end

        if episode_id
          episode = Canary::AnnictSchema.object_from_id(episode_id)

          unless episode
            raise GraphQL::ExecutionError, "Episode #{episode_id} not found"
          end

          @records = @records.joins(:episode_record).where(episode_records: {episode_id: episode.id})
        end

        if by_following
          @records = @records.joins(:user).merge(context[:viewer].followings)
        end

        @records = case order.field
        when :created_at
          @records.order(order.field => order.direction)
        when :likes_count
          @records
            .joins(:episode_record)
            .order("episode_records.likes_count": order.direction, "records.created_at": :desc)
        when :rating
          order_sql = <<~SQL
            CASE
              WHEN "episode_records"."rating_state" = 'bad' THEN '0'
              WHEN "episode_records"."rating_state" = 'average' THEN '1'
              WHEN "episode_records"."rating_state" = 'good' THEN '2'
              WHEN "episode_records"."rating_state" = 'great' THEN '3'
            END #{order.direction.upcase} NULLS LAST
          SQL

          @records
            .joins(:episode_record)
            .order(order_sql)
            .order("records.created_at": :desc)
        else
          @records
        end

        @records
      end
    end
  end
end
