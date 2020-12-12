# frozen_string_literal: true

class ApplicationResult < Dry::Struct
  schema schema.strict

  module Types
    include Dry.Types(default: :strict)

    ServiceErrorCode = Types::String.enum("error")
  end

  class BaseError < Dry::Struct
    attribute :message, Types::String
  end

  class ServiceError < BaseError
    attribute :code, Types::ServiceErrorCode
  end
end
