# frozen_string_literal: true

module Api
  module V1
    class StaffsIndexParams
      include ActiveParameter

      param :fields
      param :filter_ids
      param :filter_work_id
      param :page, default: 1
      param :per_page, default: 25
      param :sort_id
      param :sort_sort_number

      validates :fields,
        allow_blank: true,
        fields_params: true
      validates :filter_ids,
        allow_blank: true,
        filter_ids_params: true
      validates :filter_work_id,
        allow_blank: true,
        numericality: {
          only_integer: true,
          greater_than_or_equal_to: 1
        }
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
      validates :sort_sort_number,
        allow_blank: true,
        sort_params: true
    end
  end
end
