# typed: false
# frozen_string_literal: true

module Api
  module V1
    module Me
      class WorksIndexParams
        include ActiveParameter

        KINDS = /\A(wanna_watch|watching|watched|on_hold|stop_watching)\z/

        param :fields
        param :filter_ids
        param :filter_season
        param :filter_title
        param :filter_status
        param :page, default: 1
        param :per_page, default: 25
        param :sort_id
        param :sort_season
        param :sort_watchers_count

        validates :fields,
          allow_blank: true,
          fields_params: true
        validates :filter_ids,
          allow_blank: true,
          filter_ids_params: true
        validates :filter_season,
          allow_blank: true,
          filter_season_params: true
        validates :filter_status,
          allow_blank: true,
          format: {with: KINDS, message: "の値が不正です。"}
        validates :per_page,
          allow_blank: true,
          numericality: {
            only_integer: true,
            greater_than_or_equal_to: 1,
            less_than_or_equal_to: 50
          }
        validates :page,
          allow_blank: true,
          numericality: {
            only_integer: true,
            greater_than_or_equal_to: 1
          }
        validates :sort_id,
          allow_blank: true,
          sort_params: true
        validates :sort_season,
          allow_blank: true,
          sort_params: true
        validates :sort_watchers_count,
          allow_blank: true,
          sort_params: true
      end
    end
  end
end
