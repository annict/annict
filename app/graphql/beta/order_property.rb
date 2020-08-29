# frozen_string_literal: true

module Beta
  class OrderProperty
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
