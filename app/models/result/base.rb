# frozen_string_literal: true

module Result
  class Base < Dry::Struct
    schema schema.strict

    module Types
      include Dry.Types(default: :strict)
    end

    class BaseError < Dry::Struct
      attribute :message, Types::String
    end
  end
end
