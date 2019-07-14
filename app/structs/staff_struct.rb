# frozen_string_literal: true

class StaffStruct < ApplicationStruct
  attribute :name, StructTypes::Strict::String
  attribute :name_en, StructTypes::Strict::String
  attribute :resource, PersonStruct | OrganizationStruct
  attribute :resource_typename, StructTypes::Strict::String
  attribute :role, StructTypes::Strict::String
  attribute :role_en, StructTypes::Strict::String
end
