# frozen_string_literal: true

class StaffEntity < ApplicationEntity
  local_attributes :accurate_name, :role

  attribute? :accurate_name, Types::String
  attribute? :accurate_name_en, Types::String.optional
  attribute? :role, Types::String
  attribute? :role_en, Types::String.optional
  attribute? :resource do
    attribute :typename, Types::String
    attribute :id, Types::Integer
  end

  def person?
    resource.typename == "Person"
  end
end
