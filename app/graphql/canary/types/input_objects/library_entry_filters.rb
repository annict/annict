# frozen_string_literal: true

module Canary
  module Types
    module InputObjects
      class LibraryEntryFilters < Canary::Types::InputObjects::Base
        argument :status_kinds, [Canary::Types::Enums::StatusKind], required: false
      end
    end
  end
end
