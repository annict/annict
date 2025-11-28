# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Unions
      class StaffResourceItem < Canary::Types::Unions::Base
        possible_types Canary::Types::Objects::PersonType, Canary::Types::Objects::OrganizationType
      end
    end
  end
end
