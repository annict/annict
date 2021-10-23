# frozen_string_literal: true

module Beta
  module Types
    module InputObjects
      class LibraryEntryOrder < Beta::Types::InputObjects::Base
        argument :field, Beta::Types::Enums::LibraryEntryOrderField, required: true
        argument :direction, Beta::Types::Enums::OrderDirection, required: true
      end
    end
  end
end
