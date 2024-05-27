# typed: false
# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsDestroyParams
        include ActiveParameter

        param :id

        validates :id,
          presence: true,
          numericality: {
            only_integer: true,
            greater_than_or_equal_to: 1
          }
      end
    end
  end
end
