# frozen_string_literal: true

class OrganizationStruct < ApplicationStruct
  attribute :annict_id, StructTypes::Strict::Integer
  attribute :name, StructTypes::Strict::String
  attribute :name_en, StructTypes::Strict::String
end
