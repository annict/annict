# frozen_string_literal: true

class BaseServiceResult < ApplicationResult
  attribute :errors, Types::Array.of(ServiceError)
end
