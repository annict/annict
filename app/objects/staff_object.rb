# frozen_string_literal: true

class StaffObject < ApplicationObject
  attribute :person, PersonObject.optional
  attribute :organization, OrganizationObject.optional
  attribute :role_text, ObjectTypes::Strict::String
end
