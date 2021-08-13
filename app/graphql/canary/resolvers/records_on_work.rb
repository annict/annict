# frozen_string_literal: true

module Canary
  module Resolvers
    class RecordsOnWork < Canary::Resolvers::Base
      def resolve(has_body: nil, by_viewer: nil, by_following: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)

        @records = object.records.with_work_record.only_kept

        if by_viewer
          @records = @records.where(user_id: context[:viewer].id)
        end

        if by_following
          @records = @records.joins(:user).merge(context[:viewer].followings)
        end

        if has_body
          @records = @records.merge(WorkRecord.with_body)
        end

        @records = case order.field
        when :created_at
          @records.order(order.field => order.direction)
        when :likes_count
          @records
            .joins(:work_record)
            .order("work_records.likes_count": order.direction, "records.created_at": :desc)
        when :rating
          order_sql = <<~SQL
            CASE
              WHEN "work_records"."rating_overall_state" = 'bad' THEN '0'
              WHEN "work_records"."rating_overall_state" = 'average' THEN '1'
              WHEN "work_records"."rating_overall_state" = 'good' THEN '2'
              WHEN "work_records"."rating_overall_state" = 'great' THEN '3'
            END #{order.direction.upcase} NULLS LAST
          SQL

          @records
            .joins(:work_record)
            .order(Arel.sql(order_sql))
            .order("records.created_at": :desc)
        else
          @records
        end

        @records
      end
    end
  end
end
