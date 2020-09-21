# frozen_string_literal: true

module Canary
  module Resolvers
    class Records < Canary::Resolvers::Base
      def resolve(has_body: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)

        @records = object.records.only_kept

        if has_body
          @records = @records.merge(EpisodeRecord.with_body)
        end

        @records = case order.field
        when :created_at
          @records.order(order.field => order.direction)
        when :likes_count
          @records.
            joins(:episode_record).
            order("episode_records.likes_count": order.direction, "records.created_at": :desc)
        when :rating
          order_sql = <<~SQL
            CASE
              WHEN "episode_records"."rating_state" = 'bad' THEN '0'
              WHEN "episode_records"."rating_state" = 'average' THEN '1'
              WHEN "episode_records"."rating_state" = 'good' THEN '2'
              WHEN "episode_records"."rating_state" = 'great' THEN '3'
            END #{order.direction.upcase} NULLS LAST
          SQL

          @records.
            joins(:episode_record).
            order(order_sql).
            order("records.created_at": :desc)
        else
          @records
        end

        @records
      end
    end
  end
end
