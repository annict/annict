# frozen_string_literal: true

class ServiceResult < ApplicationResult
  attribute :errors, Types::Array.of(ServiceError)
end
