# frozen_string_literal: true

module Canary
  module Resolvers
    class AnimeList < Canary::Resolvers::Base
      def resolve(database_ids: nil, seasons: nil, titles: nil, order_by: nil)
        order = Canary::OrderProperty.build(order_by)

        @anime_list = Anime.only_kept

        if database_ids
          @anime_list = @anime_list.where(id: database_ids)
        end

        if seasons
          @anime_list = @anime_list.by_seasons(seasons)
        end

        if titles
          @anime_list = @anime_list.ransack(
            title_or_title_en_or_title_kana_or_title_alter_or_title_alter_en_cont_any: titles
          ).result
        end

        @anime_list = case order.field
        when :created_at, :watchers_count
          @anime_list.order(order.field => order.direction)
        when :season
          @anime_list.order_by_season(order.direction)
        else
          @anime_list
        end

        @anime_list
      end
    end
  end
end
