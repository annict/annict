# frozen_string_literal: true

module Api
  module V1
    module Me
      class RecordsUpdateParams
        include ActiveParameter

        param :id
        param :comment
        param :rating
        param :share_twitter, default: "false"
        param :share_facebook, default: "false"

        validates :id,
          presence: true,
          numericality: {
            only_integer: true,
            greater_than_or_equal_to: 1
          }
        validates :share_twitter,
          allow_blank: true,
          format: { with: /\A(true|false)\z/, message: "の値が不正です。" }
        validates :share_facebook,
          allow_blank: true,
          format: { with: /\A(true|false)\z/, message: "の値が不正です。" }
      end
    end
  end
end
