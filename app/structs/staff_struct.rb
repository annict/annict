# frozen_string_literal: true

class StaffStruct < ApplicationStruct
  attribute :person, PersonStruct
  attribute :organization, OrganizationStruct
  attribute :role_text, StructTypes::Strict::String
end
