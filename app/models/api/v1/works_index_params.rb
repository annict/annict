# frozen_string_literal: true

module Api
  module V1
    class WorksIndexParams
      include ActiveParameter

      param :fields
      param :per_page, default: 25
      param :page, default: 1
      param :sort

      validates :per_page,
        allow_blank: true,
        numericality: {
          only_integer: true,
          greater_than_or_equal_to: 1,
          less_than_or_equal_to: 100
        }
    end
  end
end
