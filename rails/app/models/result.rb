# typed: false
# frozen_string_literal: true

class Result < Dry::Struct
  schema schema.strict

  module Types
    T.unsafe(self).include Dry.Types(default: :strict)
  end

  class Error < Dry::Struct
    attribute :message, Types::String
  end

  attribute :errors, Types::Array.of(Error)

  def success?
    errors.empty?
  end
end
