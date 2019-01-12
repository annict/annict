# frozen_string_literal: true

module Types
  module InputObjects
    class ReviewOrder < Types::InputObjects::Base
      argument :field, Types::Enums::ReviewOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
