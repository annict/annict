# frozen_string_literal: true

module Api
  module V1
    class FollowingIndexParams
      include ActiveParameter

      param :fields
      param :filter_user_id
      param :filter_username
      param :page, default: 1
      param :per_page, default: 25
      param :sort_id

      validates :fields,
        allow_blank: true,
        fields_params: true
      validates :filter_user_id,
        allow_blank: true,
        numericality: {
          only_integer: true,
          greater_than_or_equal_to: 1
        }
      validates :filter_username,
        allow_blank: true,
        filter_username_params: true
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
    end
  end
end
