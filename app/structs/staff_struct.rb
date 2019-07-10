# frozen_string_literal: true

class StaffStruct < ApplicationStruct
  attribute :person, PersonStruct.optional
  attribute :organization, OrganizationStruct.optional
  attribute :role_text, StructTypes::Strict::String
end
