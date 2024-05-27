# typed: false
# frozen_string_literal: true

module Api
  module V1
    module Me
      class SlotsIndexParams
        include ActiveParameter

        param :fields
        param :filter_ids
        param :filter_channel_ids
        param :filter_work_ids
        param :filter_started_at_gt
        param :filter_started_at_lt
        param :filter_unwatched
        param :filter_rebroadcast
        param :page, default: 1
        param :per_page, default: 25
        param :sort_id
        param :sort_started_at

        validates :fields,
          allow_blank: true,
          fields_params: true
        validates :filter_ids,
          allow_blank: true,
          filter_ids_params: true
        validates :filter_channel_ids,
          allow_blank: true,
          filter_ids_params: true
        validates :filter_work_ids,
          allow_blank: true,
          filter_ids_params: true
        validates :filter_started_at_gt,
          allow_blank: true,
          filter_date_params: true
        validates :filter_started_at_lt,
          allow_blank: true,
          filter_date_params: true
        validates :filter_unwatched,
          allow_blank: true,
          filter_boolean_params: true
        validates :filter_rebroadcast,
          allow_blank: true,
          filter_boolean_params: true
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
        validates :sort_started_at,
          allow_blank: true,
          sort_params: true
      end
    end
  end
end
