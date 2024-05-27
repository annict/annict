# typed: false
# frozen_string_literal: true

module Api
  module V1
    module Me
      class StatusesCreateParams
        include ActiveParameter

        KINDS = /\A(wanna_watch|watching|watched|on_hold|stop_watching|no_select)\z/

        param :work_id
        param :kind

        validates :work_id,
          presence: true,
          numericality: {
            only_integer: true,
            greater_than_or_equal_to: 1
          }
        validates :kind,
          presence: true,
          format: {with: KINDS, message: "の値が不正です。"}
      end
    end
  end
end
