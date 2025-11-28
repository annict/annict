# typed: false
# frozen_string_literal: true

module Api
  module V1
    class RecordsIndexParams
      include ActiveParameter

      param :fields
      param :filter_ids
      param :filter_episode_id
      param :filter_has_record_comment
      param :page, default: 1
      param :per_page, default: 25
      param :sort_id
      param :sort_likes_count

      validates :fields,
        allow_blank: true,
        fields_params: true
      validates :filter_ids,
        allow_blank: true,
        filter_ids_params: true
      validates :filter_episode_id,
        allow_blank: true,
        numericality: {
          only_integer: true,
          greater_than_or_equal_to: 1
        }
      validates :filter_has_record_comment,
        allow_blank: true,
        filter_boolean_params: true
      validates :page,
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
      validates :sort_id,
        allow_blank: true,
        sort_params: true
      validates :sort_likes_count,
        allow_blank: true,
        sort_params: true
    end
  end
end
