# frozen_string_literal: true

module Types
  module InputObjects
    class OrganizationOrder < Types::InputObjects::Base
      argument :field, Types::Enums::OrganizationOrderField, required: true
      argument :direction, Types::Enums::OrderDirection, required: true
    end
  end
end
