# frozen_string_literal: true

module Types
  module Unions
    class StaffResourceItem < Types::Unions::Base
      possible_types Types::Objects::PersonType, Types::Objects::OrganizationType
    end
  end
end
