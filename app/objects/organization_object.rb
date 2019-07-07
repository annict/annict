# frozen_string_literal: true

class OrganizationObject < ApplicationObject
  attribute :annict_id, ObjectTypes::Strict::Integer
  attribute :name, ObjectTypes::Strict::String
end
