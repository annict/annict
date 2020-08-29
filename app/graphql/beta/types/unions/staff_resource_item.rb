# frozen_string_literal: true

module Beta
  module Types
    module Unions
      class StaffResourceItem < Beta::Types::Unions::Base
        possible_types Beta::Types::Objects::PersonType, Beta::Types::Objects::OrganizationType
      end
    end
  end
end
