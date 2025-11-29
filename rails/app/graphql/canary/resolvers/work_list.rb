# typed: false
# frozen_string_literal: true

module Canary
  module Resolvers
    class WorkList < Canary::Resolvers::Base
      def resolve(database_ids: nil, seasons: nil, titles: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)

        @work_list = Work.only_kept

        if database_ids
          @work_list = @work_list.where(id: database_ids)
        end

        if seasons
          @work_list = @work_list.by_seasons(seasons)
        end

        if titles
          @work_list = @work_list.ransack(
            title_or_title_en_or_title_kana_or_title_alter_or_title_alter_en_cont_any: titles
          ).result
        end

        @work_list = case order.field
        when :created_at, :watchers_count
          @work_list.order(order.field => order.direction)
        when :season
          @work_list.order_by_season(order.direction)
        else
          @work_list
        end

        @work_list
      end
    end
  end
end
