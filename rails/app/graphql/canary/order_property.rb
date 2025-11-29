# typed: false
# frozen_string_literal: true

module Canary
  class OrderProperty
    def self.build(order_by)
      return new unless order_by

      new(order_by[:field], order_by[:direction])
    end

    def initialize(field_ = nil, direction_ = nil)
      @field_ = field_
      @direction_ = direction_
    end

    def field
      field_&.to_s&.downcase&.to_sym.presence || :created_at
    end

    def direction
      direction_&.to_s&.downcase&.to_sym.presence || :asc
    end

    private

    attr_reader :field_, :direction_
  end
end
