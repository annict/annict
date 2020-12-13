# frozen_string_literal: true

module Result
  class Service < Base
    ErrorCode = Types::String.enum("error")

    class ServiceError < BaseError
      attribute :code, ErrorCode
    end

    attribute :errors, Types::Array.of(ServiceError)
  end
end
